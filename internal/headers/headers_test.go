package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParser(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFoo: Barbar\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	val, ok := headers.Get("HOsT")
	assert.True(t, ok)
	assert.Equal(t, "localhost:42069", val)
	val, ok = headers.Get("MIssing KEY")
	assert.False(t, ok)
	assert.Equal(t, "", val)
	assert.Equal(t, 36, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("Bhuvan: he is   \r\nBhuvan: the fucking\r\nBhuvan:GOAT\r\n\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	val, ok = headers.Get("BHUvan")
	assert.True(t, ok)
	assert.Equal(t, "he is,the fucking,GOAT", val)
	assert.True(t, done)
}
