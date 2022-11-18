package firetail

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"firetail-lambda-extension/logsapi"
	"testing"
	"time"

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

func TestExtractSingleRecord(t *testing.T) {
	testRecord := Record{
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)
	encodedRecord := "firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	testMessage := logsapi.LogMessage{
		Time:   time.Now().Format("2006-01-02T15:04:05.000Z"),
		Type:   "function",
		Record: json.RawMessage{},
	}
	testRecordBytes, err := json.Marshal(encodedRecord)
	require.Nil(t, err)
	testMessage.Record = testRecordBytes

	decodedRecords, err := ExtractFiretailRecords([]logsapi.LogMessage{testMessage})
	require.Nil(t, err)

	require.Len(t, decodedRecords, 1)
	assert.Equal(t, testRecord.Response, decodedRecords[0].Response)
}

func TestExtractRecordOfInvalidType(t *testing.T) {
	testMessage := logsapi.LogMessage{
		Type: "platform.start",
	}

	decodedRecords, err := ExtractFiretailRecords([]logsapi.LogMessage{testMessage})
	require.NotNil(t, err)

	assert.Len(t, decodedRecords, 0)
	assert.Contains(t, err.Error(), "logMessage type is 'platform.start', not 'function'")
}

func TestExtractRecordWithInvalidType(t *testing.T) {
	invalidRecord := 3.14159
	testMessage := logsapi.LogMessage{
		Time:   time.Now().Format("2006-01-02T15:04:05.000Z"),
		Type:   "function",
		Record: json.RawMessage{},
	}
	testRecordBytes, err := json.Marshal(invalidRecord)
	require.Nil(t, err)
	testMessage.Record = testRecordBytes

	decodedRecords, err := ExtractFiretailRecords([]logsapi.LogMessage{testMessage})
	require.NotNil(t, err)

	assert.Len(t, decodedRecords, 0)
	assert.Contains(t, err.Error(), "Err unmarshalling event record as string, err: json: cannot unmarshal number into Go value of type string")
	assert.Contains(t, err.Error(), "Err decoding event record as firetail event, err: record had 1 parts when split by ':'")
}
