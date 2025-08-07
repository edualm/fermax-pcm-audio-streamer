package config

import (
	"flag"
	"log"
)

type Config struct {
	ServerPort    int
	PythonURL     string
	AudioFilesDir string
}

func Load() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.ServerPort, "port", 8556, "HTTP server port")
	flag.StringVar(&cfg.PythonURL, "python-url", "http://127.0.0.1:8080/audio_input", "Python endpoint URL")
	flag.StringVar(&cfg.AudioFilesDir, "audio-dir", "./audio_files", "Directory containing audio files")
	flag.Parse()

	log.Printf("Config: Server port=%d, Python URL=%s, Audio dir=%s",
		cfg.ServerPort, cfg.PythonURL, cfg.AudioFilesDir)

	return cfg
}
