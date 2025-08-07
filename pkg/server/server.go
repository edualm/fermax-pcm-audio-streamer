package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"pcm-audio-streamer/pkg/audio"
	"pcm-audio-streamer/pkg/config"
)

type Server struct {
	config *config.Config
	client *http.Client
}

func New(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
		client: &http.Client{},
	}
}

func (s *Server) handlePlayFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	
	filename := string(body)
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}
	
	log.Printf("Playing file: %s", filename)
	
	pcmData, err := audio.ReadWAVFile(s.config.AudioFilesDir, filename)
	if err != nil {
		log.Printf("Error reading WAV file: %v", err)
		if err.Error() == "path traversal detected" || err.Error() == "path separators not allowed" {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
		} else {
			http.Error(w, "File not found or invalid", http.StatusNotFound)
		}
		return
	}
	
	if err := s.forwardAudioToPython(pcmData); err != nil {
		log.Printf("Error forwarding audio to Python: %v", err)
		http.Error(w, "Failed to forward audio", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Audio forwarded successfully"))
}

func (s *Server) forwardAudioToPython(audioData []byte) error {
	req, err := http.NewRequest("POST", s.config.PythonURL, bytes.NewReader(audioData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "audio/pcm")
	req.Header.Set("Content-Length", strconv.Itoa(len(audioData)))
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Python endpoint returned status %d", resp.StatusCode)
	}
	
	log.Printf("Successfully forwarded %d bytes to Python endpoint", len(audioData))
	return nil
}

func (s *Server) Start() error {
	http.HandleFunc("/play_file", s.handlePlayFile)
	
	addr := fmt.Sprintf(":%d", s.config.ServerPort)
	log.Printf("Starting server on %s", addr)
	
	return http.ListenAndServe(addr, nil)
}