package logsapi

import (
	"bytes"
	"context"
	"firetail-lambda-extension/firetail"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read failed")
}

func TestLogsApiHandlerUnreadableRequestBody(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"success"}`)
	}))
	defer testServer.Close()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:])
	errs := []error{}
	client, err := NewClient(Options{
		BatchCallback: func(batch []firetail.Record) error {
			return nil
		},
		ErrCallback: func(err error) {
			errs = append(errs, err)
		},
	})
	defer client.Shutdown(context.Background())
	require.Nil(t, err)
	require.NotNil(t, client)

	request := httptest.NewRequest("GET", "http://127.0.0.1:0", errReader(0))
	recorder := httptest.NewRecorder()
	client.logsApiHandler(recorder, request)

	require.Len(t, errs, 1)
	assert.Equal(t, "Error reading body:: read failed", errs[0].Error())
}

func TestLogsApiHandlerBadRequestBody(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"success"}`)
	}))
	defer testServer.Close()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:])
	errs := []error{}
	client, err := NewClient(Options{
		BatchCallback: func(batch []firetail.Record) error {
			return nil
		},
		ErrCallback: func(err error) {
			errs = append(errs, err)
		},
	})
	defer client.Shutdown(context.Background())
	require.Nil(t, err)
	require.NotNil(t, client)

	request := httptest.NewRequest("GET", "http://127.0.0.1:0", bytes.NewBuffer(
		[]byte(`{"executionTime":"long"}`),
	))
	recorder := httptest.NewRecorder()
	client.logsApiHandler(recorder, request)

	require.Len(t, errs, 1)
	assert.Equal(t, "Err unmarshalling Lambda Logs API request body into []LogMessage: json: cannot unmarshal object into Go value of type []logsapi.logMessage", errs[0].Error())
}
