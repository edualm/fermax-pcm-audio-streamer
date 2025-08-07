package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"pcm-audio-streamer/pkg/config"
	"pcm-audio-streamer/pkg/server"
)

func main() {
	cfg := config.Load()
	
	srv := server.New(cfg)
	
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("Shutting down server...")
}