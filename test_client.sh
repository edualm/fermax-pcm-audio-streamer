#!/bin/bash

echo "Testing PCM Audio Streamer"
echo "=========================="

echo "1. Testing valid WAV file..."
curl -X POST -d "test_tone.wav" http://localhost:8081/play_file

echo -e "\n\n2. Testing invalid filename (path traversal)..."
curl -X POST -d "../test.wav" http://localhost:8081/play_file

echo -e "\n\n3. Testing non-existent file..."
curl -X POST -d "nonexistent.wav" http://localhost:8081/play_file

echo -e "\n\n4. Testing empty filename..."
curl -X POST -d "" http://localhost:8081/play_file

echo -e "\n\n5. Testing non-WAV file..."
curl -X POST -d "test.txt" http://localhost:8081/play_file

echo -e "\n\nTest completed."