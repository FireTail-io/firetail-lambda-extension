package firetail

import (
	"encoding/json"
	"fmt"
	"io"
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
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
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
			Headers: map[string]string{
				"Test-Header-Name": "Test-Header-Value",
			},
		},
		ExecutionTime: 50,
	}
}

func getTestServer(t *testing.T, body *[]byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"success"}`)
		var err error
		*body, err = io.ReadAll(r.Body)
		assert.Nil(t, err)
	}))
}

func getBrokenTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"failure"}`)
	}))
}

func TestSendRecordToSaas(t *testing.T) {
	var receivedBody []byte
	testServer := getTestServer(t, &receivedBody)
	defer testServer.Close()

	testRecord := getValidRecord(t)

	recordsSent, err := SendRecordsToSaaS([]Record{testRecord}, testServer.URL, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, recordsSent)
	assert.Equal(t,
		"{\"dateCreated\":1668685315222,\"executionTime\":50,\"request\":{\"body\":\"\",\"headers\":{\"Content-Length\":[\"0\"],\"Host\":[\"5iagptskg6.execute-api.eu-west-2.amazonaws.com\"],\"Postman-Token\":[\"8639a798-d0e7-420a-bd98-0c5cb16c6115\"],\"User-Agent\":[\"PostmanRuntime/7.28.4\"],\"X-Amzn-Trace-Id\":[\"Root=1-63761e03-7bc79fb21f90dbbe66feba18\"],\"X-Forwarded-For\":[\"37.228.214.117\"],\"X-Forwarded-Port\":[\"443\"],\"X-Forwarded-Proto\":[\"https\"],\"accept\":[\"*/*\"],\"accept-encoding\":[\"gzip, deflate, br\"]},\"httpProtocol\":\"HTTP/1.1\",\"ip\":\"37.228.214.117\",\"method\":\"GET\",\"uri\":\"https://5iagptskg6.execute-api.eu-west-2.amazonaws.com/hi\",\"resource\":\"/hi\"},\"response\":{\"body\":\"{\\\"Description\\\":\\\"This is a test response body\\\"}\",\"headers\":{\"Test-Header-Name\":[\"Test-Header-Value\"]},\"statusCode\":200},\"version\":\"1.0.0-alpha\",\"metadata\":{\"source\":\"lambda-extension\"}}\n",
		string(receivedBody),
	)
}

func TestSendInvalidRecordToSaas(t *testing.T) {
	var receivedBody []byte
	testServer := getTestServer(t, &receivedBody)
	defer testServer.Close()

	invalidRecord := getInvalidRecord(t)

	recordsSent, err := SendRecordsToSaaS([]Record{invalidRecord}, testServer.URL, "")
	assert.Equal(t, 0, recordsSent)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "json: cannot unmarshal string into Go struct field APIGatewayProxyRequest.headers of type map[string]string")
	assert.Contains(t, err.Error(), "json: cannot unmarshal string into Go struct field APIGatewayV2HTTPRequest.headers of type map[string]string")
	assert.Nil(t, receivedBody)
}

func TestSendRecordToInvalidApiUrl(t *testing.T) {
	recordsSent, err := SendRecordsToSaaS([]Record{getValidRecord(t)}, "\n", "")
	assert.Equal(t, 0, recordsSent)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), `parse "\n": net/url: invalid control character in URL`)
}

func TestSendRecordToUnavailableSaas(t *testing.T) {
	var receivedBody []byte
	testServer := getTestServer(t, &receivedBody)
	testServer.Close()

	testRecord := getValidRecord(t)

	recordsSent, err := SendRecordsToSaaS([]Record{testRecord}, testServer.URL, "")
	assert.Equal(t, 1, recordsSent)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
	assert.Contains(t, err.Error(), "Failed to make log request, err: Post")
	assert.Nil(t, receivedBody)
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
