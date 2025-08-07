#!/usr/bin/env python3
"""
Simple test server to act as the Python audio endpoint for testing.
Receives PCM audio data and logs information about it.
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import sys

class AudioHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/audio_input':
            content_length = int(self.headers.get('Content-Length', 0))
            content_type = self.headers.get('Content-Type', '')
            
            audio_data = self.rfile.read(content_length)
            
            print(f"Received audio data:")
            print(f"  Content-Type: {content_type}")
            print(f"  Content-Length: {content_length}")
            print(f"  Data size: {len(audio_data)} bytes")
            
            self.send_response(200)
            self.send_header('Content-Type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'Audio received successfully')
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        # Suppress default HTTP log messages
        pass

if __name__ == '__main__':
    server = HTTPServer(('127.0.0.1', 8080), AudioHandler)
    print("Python test server listening on http://127.0.0.1:8080")
    print("Waiting for audio data on /audio_input endpoint...")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down test server...")
        server.shutdown()