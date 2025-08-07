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
	
	wavInfo, err := audio.ReadWAVFile(s.config.AudioFilesDir, filename)
	if err != nil {
		log.Printf("Error reading WAV file: %v", err)
		if err.Error() == "path traversal detected" || err.Error() == "path separators not allowed" {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
		} else {
			http.Error(w, "File not found or invalid", http.StatusNotFound)
		}
		return
	}
	
	// Log WAV file info
	log.Printf("WAV Info: %dHz, %d channels, %d bits, %d bytes", 
		wavInfo.SampleRate, wavInfo.Channels, wavInfo.BitsPerSample, len(wavInfo.PCMData))
	
	// Process audio data for ONVIF compatibility
	processedData, contentType, err := s.processAudioForONVIF(wavInfo)
	if err != nil {
		log.Printf("Error processing audio: %v", err)
		http.Error(w, "Audio processing failed", http.StatusInternalServerError)
		return
	}
	
	if err := s.forwardAudioToPython(processedData, contentType); err != nil {
		log.Printf("Error forwarding audio to Python: %v", err)
		http.Error(w, "Failed to forward audio", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Audio forwarded successfully"))
}

func (s *Server) processAudioForONVIF(wavInfo *audio.WAVInfo) ([]byte, string, error) {
	pcmData := wavInfo.PCMData
	
	// Check if mono
	if wavInfo.Channels != 1 {
		return nil, "", fmt.Errorf("only mono audio supported, got %d channels", wavInfo.Channels)
	}
	
	// Check if 16-bit
	if wavInfo.BitsPerSample != 16 {
		return nil, "", fmt.Errorf("only 16-bit audio supported, got %d bits", wavInfo.BitsPerSample)
	}
	
	// Resample to 8kHz if needed
	if wavInfo.SampleRate != 8000 {
		log.Printf("Resampling from %dHz to 8kHz", wavInfo.SampleRate)
		resampled, err := audio.ResamplePCMTo8kHz(pcmData, int(wavInfo.SampleRate))
		if err != nil {
			return nil, "", fmt.Errorf("resampling failed: %w", err)
		}
		pcmData = resampled
	}
	
	// Convert PCM to G.711 A-law for ONVIF compatibility
	g711Data, err := audio.ConvertPCMToG711(pcmData)
	if err != nil {
		return nil, "", fmt.Errorf("G.711 conversion failed: %w", err)
	}
	
	log.Printf("Audio processed: %d PCM bytes -> %d G.711 bytes", len(pcmData), len(g711Data))
	
	return g711Data, "audio/g711", nil
}

func (s *Server) forwardAudioToPython(audioData []byte, contentType string) error {
	req, err := http.NewRequest("POST", s.config.PythonURL, bytes.NewReader(audioData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", contentType)
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