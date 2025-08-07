# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based HTTP audio streaming server that accepts POST requests with filenames and streams corresponding audio files to a Python audio processing endpoint. The server reads audio files from a local directory and forwards them as PCM audio data.

## Architecture

The project follows a simple HTTP server pattern:
- HTTP server listening on configurable port (default: 8081)  
- POST /play_file endpoint that accepts filename in request body
- Reads audio files from ./audio_files/ subdirectory
- Streams audio data to Python endpoint at http://127.0.0.1:8080/audio_input
- Uses Content-Type: audio/pcm header for Python integration

## Common Commands

### Build and Run
```bash
# Initialize Go module (first time setup)
go mod init pcm-audio-streamer

# Build the application  
go build -o pcm-audio-streamer

# Run the application
./pcm-audio-streamer

# Run directly with Go
go run main.go
```

### Testing
```bash
# Run tests
go test ./...

# Test with verbose output
go test -v ./...

# Test a specific package
go test ./pkg/audio
```

### Development
```bash
# Format code
go fmt ./...

# Vet code for potential issues
go vet ./...

# Install dependencies
go mod tidy

# Download dependencies
go mod download
```

## Key Implementation Patterns

### Audio File Processing
- Read WAV files from ./audio_files/ directory
- Stream audio data in chunks for memory efficiency  
- Convert to PCM format if needed
- Prevent path traversal attacks in filename validation

### HTTP Integration with Python
The server forwards audio data using this pattern:
```go
req, err := http.NewRequest("POST", pythonURL, bytes.NewReader(audioData))
req.Header.Set("Content-Type", "audio/pcm")
req.Header.Set("Content-Length", strconv.Itoa(len(audioData)))
```

### Configuration
- Command-line flags or config file support
- Configurable Python endpoint URL (default: http://127.0.0.1:8080/audio_input)
- Configurable server port (default: 8081)
- Configurable audio files directory path (default: ./audio_files/)

## Directory Structure
```
.
├── main.go              # Main HTTP server implementation
├── audio_files/         # Directory containing audio files to serve
├── pkg/
│   ├── audio/          # Audio processing utilities
│   ├── config/         # Configuration management
│   └── server/         # HTTP server components
├── go.mod              # Go module definition
└── go.sum              # Go module checksums
```

## Security Considerations
- Implement path traversal protection for filename inputs
- Validate audio file formats and sizes
- Handle HTTP errors and timeouts appropriately
- Use structured logging for security events