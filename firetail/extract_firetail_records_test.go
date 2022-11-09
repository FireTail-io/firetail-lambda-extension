package firetail

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeFiretailRecordHappy(t *testing.T) {
	testRecord := Record{}
	testPayloadBytes, err := json.Marshal(testRecord)
	require.Nil(t, err)

	encodedRecord := "firetail:test-token:" + base64.StdEncoding.EncodeToString(testPayloadBytes)

	decodedRecord, err := decodeFiretailRecord(encodedRecord, "test-token")
	require.Nil(t, err)

	assert.Equal(t, *decodedRecord, testRecord)
}
