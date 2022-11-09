package firetail

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeFiretailRecordResponse(t *testing.T) {
	testRecord := Record{
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	require.Nil(t, err)

	assert.Equal(t, testRecord.Response, *&decodedRecord.Response)
}
