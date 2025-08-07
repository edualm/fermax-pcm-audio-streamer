package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type WAVHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

func ValidateFilename(filename string) error {
	if strings.Contains(filename, "..") {
		return fmt.Errorf("path traversal detected")
	}
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("path separators not allowed")
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".wav") {
		return fmt.Errorf("only WAV files are supported")
	}
	return nil
}

type WAVInfo struct {
	PCMData    []byte
	SampleRate uint32
	Channels   uint16
	BitsPerSample uint16
}

func ReadWAVFile(audioDir, filename string) (*WAVInfo, error) {
	if err := ValidateFilename(filename); err != nil {
		return nil, err
	}
	
	filepath := filepath.Join(audioDir, filename)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Read RIFF header
	var riffHeader struct {
		ChunkID   [4]byte
		ChunkSize uint32
		Format    [4]byte
	}
	
	if err := binary.Read(file, binary.LittleEndian, &riffHeader); err != nil {
		return nil, fmt.Errorf("failed to read RIFF header: %w", err)
	}
	
	if string(riffHeader.ChunkID[:]) != "RIFF" || string(riffHeader.Format[:]) != "WAVE" {
		return nil, fmt.Errorf("invalid WAV file format")
	}
	
	// Read fmt chunk
	var fmtHeader struct {
		ChunkID       [4]byte
		ChunkSize     uint32
		AudioFormat   uint16
		NumChannels   uint16
		SampleRate    uint32
		ByteRate      uint32
		BlockAlign    uint16
		BitsPerSample uint16
	}
	
	if err := binary.Read(file, binary.LittleEndian, &fmtHeader); err != nil {
		return nil, fmt.Errorf("failed to read fmt chunk: %w", err)
	}
	
	if string(fmtHeader.ChunkID[:]) != "fmt " {
		return nil, fmt.Errorf("expected fmt chunk, got %s", string(fmtHeader.ChunkID[:]))
	}
	
	if fmtHeader.AudioFormat != 1 {
		return nil, fmt.Errorf("only PCM format supported, got format %d", fmtHeader.AudioFormat)
	}
	
	// Skip any extra bytes in fmt chunk (some files have extended format info)
	if fmtHeader.ChunkSize > 16 {
		extraBytes := fmtHeader.ChunkSize - 16
		if _, err := file.Seek(int64(extraBytes), io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to skip extra fmt bytes: %w", err)
		}
	}
	
	// Find data chunk (skip any other chunks like LIST, INFO, etc.)
	var dataSize uint32
	for {
		var chunkHeader struct {
			ChunkID   [4]byte
			ChunkSize uint32
		}
		
		if err := binary.Read(file, binary.LittleEndian, &chunkHeader); err != nil {
			return nil, fmt.Errorf("failed to read chunk header: %w", err)
		}
		
		if string(chunkHeader.ChunkID[:]) == "data" {
			dataSize = chunkHeader.ChunkSize
			break
		}
		
		// Skip this chunk
		if _, err := file.Seek(int64(chunkHeader.ChunkSize), io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to skip chunk %s: %w", string(chunkHeader.ChunkID[:]), err)
		}
	}
	
	// Read PCM data
	pcmData := make([]byte, dataSize)
	if _, err := io.ReadFull(file, pcmData); err != nil {
		return nil, fmt.Errorf("failed to read PCM data: %w", err)
	}
	
	return &WAVInfo{
		PCMData:       pcmData,
		SampleRate:    fmtHeader.SampleRate,
		Channels:      fmtHeader.NumChannels,
		BitsPerSample: fmtHeader.BitsPerSample,
	}, nil
}