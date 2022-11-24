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
