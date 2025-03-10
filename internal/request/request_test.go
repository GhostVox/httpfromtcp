package request

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestRequestLineParser(t *testing.T) {
	// Define test cases
	tests := []struct {
		name            string
		input           string
		numBytesPerRead int
		expectError     bool
		method          string
		target          string
		version         string
	}{
		{
			name:            "Invalid lowercase method",
			input:           "get / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			expectError:     true,
			numBytesPerRead: 2,
		},
		{
			name:            "Valid basic GET request",
			input:           "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 6,
			expectError:     false,
			method:          "GET",
			target:          "/",
			version:         "1.1",
		},
		{
			name:            "Valid GET request with path",
			input:           "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: 8,
			expectError:     false,
			method:          "GET",
			target:          "/coffee",
			version:         "1.1",
		},
		{
			name:            "Invalid HTTP version",
			input:           "GET / HTTP/1.0\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			expectError:     true,
			numBytesPerRead: 4,
		},
		{
			name:            "Invalid request line parts",
			input:           "GET / HTTP/1.1 extra\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			expectError:     true,
			numBytesPerRead: 3,
		},
		{
			name:            "Valid Post request",
			input:           "POST / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nContent-Length: 5\r\n\r\nHello",
			numBytesPerRead: len("POST / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nContent-Length: 5\r\n\r\n"),
			expectError:     false,
			method:          "POST",
			target:          "/",
			version:         "1.1",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			reader := &chunkReader{
				data:            tc.input,
				numBytesPerRead: tc.numBytesPerRead,

				pos: 0,
			}

			r, err := RequestFromReader(reader)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)
				assert.Equal(t, tc.method, r.RequestLine.Method)
				assert.Equal(t, tc.target, r.RequestLine.RequestTarget)
				assert.Equal(t, tc.version, r.RequestLine.HttpVersion)
			}
		})
	}
}

func TestHeaders_Parse(t *testing.T) {

	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Headers))

	// Test: Duplicate Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: localhost:42070\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069, localhost:42070", r.Headers["host"])

	// Test: Case Insensitive Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nhost: localhost:42070\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069, localhost:42070", r.Headers["host"])

	// Test: Missing End of Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*",
		numBytesPerRead: 3,
		pos:             0,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}
