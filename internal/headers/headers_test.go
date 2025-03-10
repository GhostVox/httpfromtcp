package headers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHeaders_Parse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	fmt.Printf(" expect n: 23 n: %d, expect done: false done: %t, expect err: false err: %v\n", n, done, err)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra spacing
	headers = NewHeaders()
	data = []byte("      Host: localhost:42069        \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: valid 2 headers with exising headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nUser-Agent: curl/7.64.1\r\ntoken: dog\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	data = data[n:]
	x, done, err := headers.Parse(data)
	assert.Equal(t, "curl/7.64.1", headers["user-agent"])
	data = data[x:]
	y, done, err := headers.Parse(data)
	assert.Equal(t, "dog", headers["token"])
	data = data[y:]
	z, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, z)
	assert.True(t, done)

	// Test: Valid Done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Valid 2 headers with existing headers
	headers = map[string]string{"host": "localhost:42069"}
	data = []byte("User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, 25, n)
	assert.False(t, done)
	assert.False(t, headers["accept"] == "*/*")

	// Test: Valid Special characters
	headers = NewHeaders()
	data = []byte("!#$%&'*+-.^_`|~: special characters\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "special characters", headers["!#$%&'*+-.^_`|~"])

	// Test: Valid multiple value header
	headers = NewHeaders()
	data = []byte("Set-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-person: tj-loves-ocaml\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "lane-loves-go", headers["set-person"])
	x, done, err = headers.Parse(data[n:])
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	y, done, err = headers.Parse(data[n+x:])
	require.NoError(t, err)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	fmt.Printf("expect n :0 n: %d, expect done: false done: %t, expect err:true err: %v\n", n, done, err)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header key
	headers = NewHeaders()
	data = []byte("Høst: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
}
