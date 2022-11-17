package firetail

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeAndDecodeLogEntry(t *testing.T) {
	testLogEntry := LogEntry{
		DateCreated:   time.Now().UnixMilli(),
		ExecutionTime: 50,
		Request: LogEntryRequest{
			Body: "{\"Description\":\"This is a test request body\"}",
			Headers: map[string][]string{
				"Test-Request-Header": []string{"Test-Value-1", "Test-Value-2"},
			},
			HTTPProtocol: HTTP2,
			IP:           "8.8.8.8",
			Method:       Get,
			URI:          "https://www.example.com/pets/100",
			Resource:     "/pets/{id}",
		},
		Response: LogEntryResponse{
			Body: "{\"Description\":\"This is a test response body\"}",
			Headers: map[string][]string{
				"Test-Response-Header": []string{"Test-Value-1", "Test-Value-2"},
			},
			StatusCode: 200,
		},
		Version: The100Alpha,
	}

	testLogEntryBytes, err := testLogEntry.Marshal()
	require.Nil(t, err)

	unmarshalledLogEntry, err := UnmarshalLogEntry(testLogEntryBytes)
	require.Nil(t, err)
	assert.Equal(t, testLogEntry, unmarshalledLogEntry)

	remarshalledLogEntryBytes, err := unmarshalledLogEntry.Marshal()
	require.Nil(t, err)
	assert.Equal(t, testLogEntryBytes, remarshalledLogEntryBytes)
}
