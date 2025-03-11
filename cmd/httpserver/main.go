package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/GhostVox/httptcp/internal/headers"
	"github.com/GhostVox/httptcp/internal/request"
	"github.com/GhostVox/httptcp/internal/response"
	"github.com/GhostVox/httptcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}

	if strings.Contains(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
		return

	}

	handler200(w, req)
}

func handler200(w response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.Success)
	Message := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>` + "\n"
	headers := response.GetDefaultHeaders(len(Message))
	w.WriteHeaders(headers)
	w.Writer.Write([]byte(Message))
}

func handler400(w response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.BadRequest)
	Message := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>` + "\n"
	headers := response.GetDefaultHeaders(len(Message))
	w.WriteHeaders(headers)
	w.Writer.Write([]byte(Message))

}

func handler500(w response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.InternalServerError)
	Message := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>` + "\n"
	headers := response.GetDefaultHeaders(len(Message))
	w.WriteHeaders(headers)
	w.Writer.Write([]byte(Message))
}

func proxyHandler(w response.Writer, req *request.Request) {
	client := &http.Client{}
	target := req.RequestLine.RequestTarget
	// Remove the /httpbin prefix
	if strings.HasPrefix(target, "/httpbin") {
		target = strings.TrimPrefix(target, "/httpbin")
	}
	// Make request to httpbin
	requestAddr := fmt.Sprintf("http://httpbin.org%s", target)
	request, err := http.NewRequest(req.RequestLine.Method, requestAddr, nil)
	if err != nil {
		handler500(w, req)
		return
	}

	// Copy original request headers
	for key, values := range req.Headers {
		request.Header.Set(key, values)
	}

	// Send request
	resp, err := client.Do(request)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	// Write the status line with the appropriate status code

	w.WriteStatusLine(response.Success)

	// Set headers and transfer them to my server's response
	h := headers.Headers{}
	resp.Header.Set("Transfer-Encoding", "chunked")
	resp.Header.Del("Content-Length")
	resp.Header.Set("Connection", "keep-alive")
	for key, values := range resp.Header {
		for _, value := range values {
			h.Set(key, value)
		}
	}

	// Advertise that trailers will be included
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	// Create a hash instance to calculate SHA-256 on the fly
	hasher := sha256.New()

	// Store the entire content to calculate hash
	var allContent []byte

	// Make a buffer to read the response body
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// Save this chunk for later hashing
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			allContent = append(allContent, chunk...)

			// Write the chunk to the client
			_, writeErr := w.WriteChunkedBody(buf[:n])
			if writeErr != nil {
				log.Printf("Error writing chunk: %v", writeErr)

				break
			}
		}

		if err != nil {
			if err == io.EOF {
				// End of file reached
				err := w.WriteChunkedBodyEnd()
				if err != nil {
					log.Printf("Error writing chunk end: %v", err)
					break
				}
				break
			}
			log.Printf("Error reading from source: %v", err)
			return
		}
	}

	totalContent := len(allContent)
	// Calculate SHA-256 hash of the entire content
	hasher.Write(allContent)
	hash := hasher.Sum(nil)

	// Add trailers
	trailers := headers.Headers{}
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", totalContent))
	w.WriteTrailers(trailers)

	// Log verification
	log.Printf("Total content: %d bytes, SHA-256: %x", totalContent, hash)
}
