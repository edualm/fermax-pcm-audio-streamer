# PCM Audio Streamer

A Go HTTP server that accepts audio file requests and streams them as PCM data to a Python endpoint.

## Features

- HTTP server with configurable port (default: 8081)
- `/play_file` POST endpoint for audio playback requests
- WAV file support with PCM format validation
- Security protection against path traversal attacks
- Configurable Python endpoint URL
- Structured logging and error handling

## Quick Start

1. **Build the application:**
   ```bash
   go build -o pcm-audio-streamer
   ```

2. **Run the server:**
   ```bash
   ./pcm-audio-streamer
   ```

3. **Play an audio file:**
   ```bash
   curl -X POST -d "test_tone.wav" http://localhost:8081/play_file
   ```

## Configuration

The server accepts the following command-line flags:

- `-port`: HTTP server port (default: 8081)
- `-python-url`: Python endpoint URL (default: http://127.0.0.1:8080/audio_input)
- `-audio-dir`: Directory containing audio files (default: ./audio_files)

Example with custom configuration:
```bash
./pcm-audio-streamer -port 9000 -python-url http://localhost:8000/process -audio-dir /path/to/audio
```

## Testing

### Manual Testing

1. **Start the Python test server** (in one terminal):
   ```bash
   python3 test_server.py
   ```

2. **Start the audio streamer** (in another terminal):
   ```bash
   ./pcm-audio-streamer
   ```

3. **Run the test client** (in a third terminal):
   ```bash
   ./test_client.sh
   ```

### Test Files

The project includes a test WAV file (`audio_files/test_tone.wav`) - a 1-second 440Hz sine wave for testing.

## API

### POST /play_file

Plays an audio file by streaming its PCM data to the configured Python endpoint.

**Request:**
- Method: POST
- Body: filename (plain text)
- Example: `test_tone.wav`

**Response:**
- 200 OK: Audio forwarded successfully
- 400 Bad Request: Invalid filename or missing file
- 404 Not Found: File not found
- 500 Internal Server Error: Server error

## Security

- Filename validation prevents path traversal attacks
- Only WAV files are supported
- Path separators in filenames are rejected
- File access is restricted to the configured audio directory

## Architecture

```
┌─────────────────┐    POST /play_file    ┌─────────────────┐
│   HTTP Client   │─────────────────────▶│  Go HTTP Server │
└─────────────────┘                       └─────────────────┘
                                                    │
                                          1. Read WAV file
                                          2. Extract PCM data
                                          3. Forward to Python
                                                    │
                                                    ▼
                                          ┌─────────────────┐
                                          │ Python Endpoint │
                                          │  audio/pcm      │
                                          └─────────────────┘
```