package firetail

import (
	"encoding/base32"
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

func TestDecodeFiretailRecordWithMissingPart(t *testing.T) {
	encodedRecord := "firetail:log-ext"

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	assert.Nil(t, decodedRecord)
	require.NotNil(t, err)

	assert.Equal(t, "record had 2 parts when split by ':'", err.Error())
}

func TestDecodeFiretailRecordWithExtraPart(t *testing.T) {
	testRecord := Record{
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes) + ":extra"

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	assert.Nil(t, decodedRecord)
	require.NotNil(t, err)

	assert.Equal(t, "record had 4 parts when split by ':'", err.Error())
}

func TestDecodeFiretailRecordWithInvalidPrefix(t *testing.T) {
	testRecord := Record{
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "tailfire:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	assert.Nil(t, decodedRecord)
	require.NotNil(t, err)

	assert.Equal(t, "record did not have firetail prefix", err.Error())
}

func TestDecodeFiretailRecordWithInvalidToken(t *testing.T) {
	testRecord := Record{
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "firetail:ext-log:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	assert.Nil(t, decodedRecord)
	require.NotNil(t, err)

	assert.Equal(t, "firetail prefixed record did not have valid token", err.Error())
}

func TestDecodeFiretailRecordWithInvalidPayloadEncoding(t *testing.T) {
	testRecord := Record{
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "firetail:log-ext:" + base32.StdEncoding.EncodeToString(testPayloadBytes)

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	assert.Nil(t, decodedRecord)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "failed to b64 decode firetail record")
}

func TestDecodeFiretailRecordWithInvalidPayloadTypes(t *testing.T) {
	type Event struct {
		Event         string `json:"event"`
		Response      int    `json:"response"`
		ExecutionTime string `json:"execution_time"`
	}
	testRecord := Event{
		Event:         "my birthday party",
		Response:      0,
		ExecutionTime: "tomorrow",
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	assert.Nil(t, decodedRecord)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "failed to unmarshal firetail event")
}
