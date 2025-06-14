package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	headers := NewHeaders()

	// Test: Valid single header
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	data = []byte("  User-Agent:     curl/8.13.0      \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "curl/8.13.0", headers["User-Agent"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	data = []byte("Content-Type: application/json\r\n  Content-Length:  22  \r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "curl/8.13.0", headers["User-Agent"])
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "", headers["Content-Length"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test: Valid done
	data = []byte("\r\n Extra request data \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 0, n)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
