# HTTP from TCP - Custom HTTP Server Implementation

A custom HTTP/1.1 server implementation built from scratch using Go's TCP networking primitives. This project demonstrates low-level network programming by implementing the HTTP protocol directly over TCP connections without using Go's standard `net/http` package.

## 🎯 Project Overview

This project implements a complete HTTP/1.1 server that:

- Parses raw HTTP requests from TCP streams
- Handles multiple connection types and response formats
- Implements chunked transfer encoding with trailers
- Provides proxy functionality to external services
- Serves static video content with streaming capabilities

## ✨ Features

- **Raw TCP Connection Handling**: Direct TCP socket management with graceful connection handling
- **HTTP/1.1 Protocol Implementation**: Complete request/response cycle parsing and generation
- **Chunked Transfer Encoding**: Streaming responses with trailer support for metadata
- **Proxy Server**: Forward requests to external HTTP services (httpbin.org)
- **Video Streaming**: Serve MP4 content with proper headers and SHA-256 integrity checking
- **Error Handling**: Custom error pages with appropriate HTTP status codes
- **Concurrent Connections**: Goroutine-based request handling for multiple simultaneous clients

## 🚀 Quick Start

### Prerequisites

- Go 1.24.1 or higher
- Video file at `../httpfromtcp/assets/vim.mp4` (for video streaming endpoint)

### Installation & Running

1. **Clone the repository**

   ```bash
   git clone https://github.com/GhostVox/httptcp.git
   cd httptcp
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Run the HTTP server**

   ```bash
   go run cmd/httpserver/main.go
   ```

   The server will start on port `42069`

4. **Test with curl**

   ```bash
   # Basic success response
   curl http://localhost:42069/

   # Error responses
   curl http://localhost:42069/yourproblem  # 400 Bad Request
   curl http://localhost:42069/myproblem    # 500 Internal Server Error

   # Proxy to httpbin
   curl http://localhost:42069/httpbin/get

   # Video streaming
   curl http://localhost:42069/video -o output.mp4
   ```

## 📁 Project Structure

```
httptcp/
├── cmd/
│   ├── httpserver/          # Main HTTP server application
│   │   └── main.go
│   ├── tcplistener/         # TCP connection listener (debugging tool)
│   │   └── main.go
│   └── udpsender/           # UDP sender utility
│       └── main.go
├── internal/
│   ├── headers/             # HTTP header parsing and management
│   │   ├── headers.go
│   │   └── headers_test.go
│   ├── request/             # HTTP request parsing
│   │   ├── request.go
│   │   └── request_test.go
│   ├── response/            # HTTP response generation
│   │   └── response.go
│   └── server/              # TCP server and connection handling
│       └── server.go
├── assets/                  # Static assets (gitignored)
├── go.mod
├── go.sum
└── messages.txt             # Sample messages for UDP testing
```

## 🔧 Technical Implementation

### Core Components

#### 1. Request Parser (`internal/request/`)

- **Stream Processing**: Parses HTTP requests incrementally from TCP streams
- **State Machine**: Handles request line → headers → body parsing states
- **Buffer Management**: Dynamic buffer resizing for variable request sizes
- **Validation**: RFC-compliant HTTP method and version validation

```go
type Request struct {
    RequestLine RequestLine
    Headers     headers.Headers
    Body        []byte
    state       state
}
```

#### 2. Header Management (`internal/headers/`)

- **Case-Insensitive**: Proper HTTP header key normalization
- **Multi-Value Support**: Handles duplicate headers with comma separation
- **Validation**: RFC-compliant header key character validation
- **Parsing**: Incremental header parsing with CRLF detection

#### 3. Response Writer (`internal/response/`)

- **State Tracking**: Ensures proper response order (status → headers → body)
- **Chunked Encoding**: Implements HTTP/1.1 chunked transfer encoding
- **Trailer Support**: Adds metadata after response body completion
- **Multiple Formats**: Standard, chunked, and streaming response support

#### 4. Server Implementation (`internal/server/`)

- **Graceful Shutdown**: Signal handling for clean server termination
- **Concurrent Handling**: Goroutine per connection for scalability
- **Error Recovery**: Proper error responses for malformed requests
- **Connection Management**: Automatic connection cleanup and resource management

### HTTP Endpoints

| Endpoint       | Method | Description                                 | Response Type                |
| -------------- | ------ | ------------------------------------------- | ---------------------------- |
| `/`            | GET    | Success page with HTML content              | Standard                     |
| `/yourproblem` | Any    | 400 Bad Request error page                  | Standard                     |
| `/myproblem`   | Any    | 500 Internal Server Error page              | Standard                     |
| `/httpbin/*`   | Any    | Proxy to httpbin.org with path forwarding   | Chunked with trailers        |
| `/video`       | GET    | MP4 video streaming with integrity checking | Chunked with SHA-256 trailer |

### Advanced Features

#### Chunked Transfer Encoding

```go
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
    w.Writer.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
    n, err := w.Writer.Write(p)
    w.Writer.Write([]byte("\r\n"))
    return n, err
}
```

#### Content Integrity Verification

The video endpoint calculates SHA-256 hash on-the-fly during streaming:

```go
hasher := sha256.New()
// Stream content while hashing
trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
trailers.Set("X-Content-Length", fmt.Sprintf("%d", totalContent))
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/headers/
go test ./internal/request/
```

### Test Coverage

- **Header Parsing**: Valid/invalid header formats, special characters, multi-values
- **Request Parsing**: Various request formats, chunked reading, error conditions
- **Edge Cases**: Malformed requests, incomplete data, buffer overflows

## 🛠️ Development Tools

### TCP Listener (Debugging)

```bash
go run cmd/tcplistener/main.go
```

Listens on port 42069 and prints raw HTTP request details for debugging.

### UDP Sender (Testing)

```bash
go run cmd/udpsender/main.go
```

Interactive UDP message sender for network testing.

## 📊 Performance Characteristics

- **Memory Efficient**: Dynamic buffer growth prevents over-allocation
- **Streaming Support**: Low memory footprint for large file transfers
- **Concurrent**: Handles multiple connections simultaneously
- **Non-blocking**: Goroutine-based architecture prevents connection blocking

## 🔒 Security Considerations

- **Input Validation**: RFC-compliant parsing prevents malformed request attacks
- **Resource Limits**: Buffer size management prevents memory exhaustion
- **Connection Timeouts**: Prevents resource leaks from hanging connections
- **Error Isolation**: Request errors don't crash the entire server

## 🚧 Limitations & Future Improvements

### Current Limitations

- HTTP/1.1 only (no HTTP/2 support)
- Basic authentication not implemented
- No HTTPS/TLS support
- Limited caching mechanisms
- Single-threaded request processing per connection

### Potential Enhancements

- [ ] TLS/SSL encryption support
- [ ] HTTP/2 protocol implementation
- [ ] Request routing and middleware system
- [ ] Static file serving with caching
- [ ] WebSocket upgrade support
- [ ] Request rate limiting
- [ ] Comprehensive logging system

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## 📝 License

This project is available under the MIT License.

## 🙏 Acknowledgments

- HTTP/1.1 specification (RFC 7230-7235)
- Go networking documentation and examples
- httpbin.org for providing testing endpoints

---

**Note**: This project is primarily educational, demonstrating low-level network programming concepts. For production use, consider Go's standard `net/http` package which provides additional security, performance, and feature completeness.
