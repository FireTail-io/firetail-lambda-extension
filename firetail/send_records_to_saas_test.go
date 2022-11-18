package firetail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getValidRecord(t *testing.T) Record {
	apiGatewayProxyRequestBytes, err := json.Marshal(getNewAPIGatewayProxyRequest())
	require.Nil(t, err)
	return Record{
		Event: json.RawMessage(apiGatewayProxyRequestBytes),
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "{\"Description\":\"This is a test response body\"}",
		},
		ExecutionTime: 50,
	}
}

func getInvalidRecord(t *testing.T) Record {
	type InvalidPayload struct {
		Headers string
	}
	invalidPayload := InvalidPayload{
		Headers: "No headers here",
	}
	invalidPayloadBytes, err := json.Marshal(invalidPayload)
	require.Nil(t, err)
	return Record{
		Event: json.RawMessage(invalidPayloadBytes),
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "{\"Description\":\"This is a test response body\"}",
		},
		ExecutionTime: 50,
	}
}

func getTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"success"}`)
	}))
}

func getBrokenTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"failure"}`)
	}))
}

func TestSendRecordToSaas(t *testing.T) {
	testServer := getTestServer()
	defer testServer.Close()

	testRecord := getValidRecord(t)

	recordsSent, err := SendRecordsToSaaS([]Record{testRecord}, testServer.URL, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, recordsSent)
}

func TestSendInvalidRecordToSaas(t *testing.T) {
	testServer := getTestServer()
	defer testServer.Close()

	invalidRecord := getInvalidRecord(t)

	recordsSent, err := SendRecordsToSaaS([]Record{invalidRecord}, testServer.URL, "")
	assert.Equal(t, 0, recordsSent)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "json: cannot unmarshal string into Go struct field APIGatewayProxyRequest.headers of type map[string]string")
	assert.Contains(t, err.Error(), "json: cannot unmarshal string into Go struct field APIGatewayV2HTTPRequest.headers of type map[string]string")
}

func TestSendRecordToInvalidApiUrl(t *testing.T) {
	recordsSent, err := SendRecordsToSaaS([]Record{getValidRecord(t)}, "\n", "")
	assert.Equal(t, 0, recordsSent)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), `parse "\n": net/url: invalid control character in URL`)
}

func TestSendRecordToUnavailableSaas(t *testing.T) {
	testServer := getTestServer()
	testServer.Close()

	testRecord := getValidRecord(t)

	recordsSent, err := SendRecordsToSaaS([]Record{testRecord}, testServer.URL, "")
	assert.Equal(t, 0, recordsSent)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
	assert.Contains(t, err.Error(), "Failed to make log request, err: Post")
}

func TestSendRecordWithSaasFailure(t *testing.T) {
	testServer := getBrokenTestServer()
	defer testServer.Close()

	testRecord := getValidRecord(t)

	recordsSent, err := SendRecordsToSaaS([]Record{testRecord}, testServer.URL, "")
	assert.Equal(t, 1, recordsSent)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "Got err response from firetail api: map[message:failure]")
}
