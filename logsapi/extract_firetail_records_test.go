package logsapi

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"firetail-lambda-extension/firetail"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeFiretailRecordResponse(t *testing.T) {
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
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

func TestDecodeFiretailRecordWithExtraPartSuffixed(t *testing.T) {
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes) + ":extra"

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	assert.Nil(t, decodedRecord)
	require.NotNil(t, err)

	assert.Equal(t, "record did not have firetail prefix", err.Error())
}

func TestDecodeFiretailRecordWithExtraPartPrefixed(t *testing.T) {
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "extra:firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes) + ""

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	require.Nil(t, err)

	assert.Equal(t, testRecord.Response, *&decodedRecord.Response)
}

func TestDecodeFiretailRecordWithInvalidPrefix(t *testing.T) {
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
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

func TestDecodeFiretailRecordWithTimestampPrefix(t *testing.T) {
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "2023-02-09T14:12:59.574Z    7b9025e7-228f-4f39-ab16-1cadba2bb3f6    INFO    firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	decodedRecord, err := decodeFiretailRecord(encodedRecord)
	require.Nil(t, err)

	assert.Equal(t, testRecord.Response, *&decodedRecord.Response)
}

func TestDecodeFiretailRecordWithInvalidToken(t *testing.T) {
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
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
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
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
	testRecord := firetail.Record{
		Response: firetail.RecordResponse{
			StatusCode: 200,
			Body:       "Test Body",
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
		},
	}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)
	encodedRecord := "firetail:log-ext:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	testMessages := []logMessage{{
		Time:   time.Now().Format("2006-01-02T15:04:05.000Z"),
		Type:   "function",
		Record: json.RawMessage{},
	}}
	testRecordBytes, err := json.Marshal(encodedRecord)
	require.Nil(t, err)
	testMessages[0].Record = testRecordBytes
	testMessageBytes, err := json.Marshal(testMessages)
	require.Nil(t, err)

	decodedRecords, err := extractFiretailRecords(testMessageBytes)
	require.Nil(t, err)

	require.Len(t, decodedRecords, 1)
	assert.Equal(t, testRecord.Response, decodedRecords[0].Response)
}

func TestExtractRecordOfInvalidType(t *testing.T) {
	testMessages := []logMessage{{
		Type: "platform.start",
	}}
	testMessageBytes, err := json.Marshal(testMessages)
	require.Nil(t, err)

	decodedRecords, err := extractFiretailRecords(testMessageBytes)
	require.NotNil(t, err)

	assert.Len(t, decodedRecords, 0)
	assert.Contains(t, err.Error(), "logMessage type is 'platform.start', not 'function'")
}

func TestExtractRecordWithInvalidType(t *testing.T) {
	invalidRecord := 3.14159
	testMessages := []logMessage{{
		Time:   time.Now().Format("2006-01-02T15:04:05.000Z"),
		Type:   "function",
		Record: json.RawMessage{},
	}}
	testRecordBytes, err := json.Marshal(invalidRecord)
	require.Nil(t, err)
	testMessages[0].Record = testRecordBytes
	testMessageBytes, err := json.Marshal(testMessages)
	require.Nil(t, err)

	decodedRecords, err := extractFiretailRecords(testMessageBytes)
	require.NotNil(t, err)

	assert.Len(t, decodedRecords, 0)
	assert.Contains(t, err.Error(), "Err unmarshalling event record as string, err: json: cannot unmarshal number into Go value of type string")
	assert.Contains(t, err.Error(), "Err decoding event record as firetail event, err: record had 1 parts when split by ':'")
}

func TestExtractMessagesWithInvalidType(t *testing.T) {
	testMessageBytes := []byte(`{"hello":"world"}`)

	decodedRecords, err := extractFiretailRecords(testMessageBytes)
	assert.Len(t, decodedRecords, 0)
	require.NotNil(t, err)
	assert.Equal(t, "Err unmarshalling Lambda Logs API request body into []LogMessage: json: cannot unmarshal object into Go value of type []logsapi.logMessage", err.Error())
}
