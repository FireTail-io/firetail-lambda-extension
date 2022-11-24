package extensionsapi

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextEvent(t *testing.T) {
	testEventResponse := NextEventResponse{
		EventType:          "TEST_EVENT_TYPE",
		DeadlineMs:         3142,
		RequestID:          "TEST_REQUEST_ID",
		InvokedFunctionArn: "TEST_FUNCTION_ARN",
		Tracing: Tracing{
			Type:  "TEST_TRACE_TYPE",
			Value: "TEST_TRACE_VALUE",
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "", string(body))

		res, err := json.Marshal(testEventResponse)
		require.Nil(t, err)

		w.Write(res)
	}))
	defer testServer.Close()

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	ctx := context.Background()
	res, err := client.NextEvent(ctx)
	require.Nil(t, err)

	assert.Equal(t, testEventResponse, *res)
}

func TestNextEventBadUrl(t *testing.T) {
	client := NewClient("\n")
	ctx := context.Background()
	res, err := client.NextEvent(ctx)
	assert.Nil(t, res)
	require.NotNil(t, err)
	assert.Equal(t, "parse \"http://\\n/2020-01-01/extension/event/next\": net/url: invalid control character in URL", err.Error())
}

func TestNextEventNoServer(t *testing.T) {
	client := NewClient("127.0.0.1:0")
	ctx := context.Background()
	res, err := client.NextEvent(ctx)
	assert.Nil(t, res)
	require.NotNil(t, err)
	assert.Equal(t, "Get \"http://127.0.0.1:0/2020-01-01/extension/event/next\": dial tcp 127.0.0.1:0: connect: can't assign requested address", err.Error())
}

func TestNextEventInternalServerError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "", string(body))

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"internal server error"}`))
	}))
	defer testServer.Close()

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	ctx := context.Background()
	res, err := client.NextEvent(ctx)
	assert.Nil(t, res)
	require.NotNil(t, err)

	assert.Equal(t, "request failed with status 500 Internal Server Error", err.Error())
}

func TestNextEventInvalidResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "", string(body))

		w.Write([]byte(`{"eventType":3142}`))
	}))
	defer testServer.Close()

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	ctx := context.Background()
	res, err := client.NextEvent(ctx)
	assert.Nil(t, res)
	require.NotNil(t, err)

	assert.Equal(t, "json: cannot unmarshal number into Go struct field NextEventResponse.eventType of type extensionsapi.EventType", err.Error())
}
