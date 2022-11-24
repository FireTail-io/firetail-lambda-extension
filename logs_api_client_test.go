package main

import (
	"context"
	"firetail-lambda-extension/logsapi"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMockRuntimeApiServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"eventType": "SHUTDOWN"}`)
	}))
}

func TestInitLogsApiClient(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	runtimeApiServer := getMockRuntimeApiServer()
	defer runtimeApiServer.Close()

	extensionID := "TEST_EXTENSION_ID"
	ctx := context.Background()

	var closeErr error
	errWaitgroup := sync.WaitGroup{}
	errWaitgroup.Add(1)
	errCallback := func(err error) {
		closeErr = err
		errWaitgroup.Done()
	}

	logsApiClient, err := initLogsApiClient(
		logsapi.Options{
			ExtensionID:         extensionID,
			LogServerAddress:    "127.0.0.1:0",
			AwsLambdaRuntimeAPI: strings.Join(strings.Split(runtimeApiServer.URL, ":")[1:], ":")[2:],
			ErrCallback:         &errCallback,
		},
		ctx,
	)
	require.Nil(t, err)

	err = logsApiClient.Shutdown(ctx)
	require.Nil(t, err)

	errWaitgroup.Wait()
	assert.Equal(t, http.ErrServerClosed, closeErr)
}

func TestInitLogsApiClientWithInvalidAddress(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	runtimeApiServer := getMockRuntimeApiServer()
	defer runtimeApiServer.Close()

	extensionID := "TEST_EXTENSION_ID"
	ctx := context.Background()

	var closeErr error
	errWaitgroup := sync.WaitGroup{}
	errWaitgroup.Add(1)
	errCallback := func(err error) {
		closeErr = err
		errWaitgroup.Done()
	}

	_, err := initLogsApiClient(
		logsapi.Options{
			ExtensionID:         extensionID,
			LogServerAddress:    ":::",
			AwsLambdaRuntimeAPI: strings.Join(strings.Split(runtimeApiServer.URL, ":")[1:], ":")[2:],
			ErrCallback:         &errCallback,
		},
		ctx,
	)
	require.Nil(t, err)

	errWaitgroup.Wait()
	assert.Equal(t, "Log server closed unexpectedly: listen tcp: address :::: too many colons in address", closeErr.Error())
}

func TestInitLogsApiClientWithNoRuntimeApi(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)

	extensionID := "TEST_EXTENSION_ID"
	ctx := context.Background()

	_, err := initLogsApiClient(
		logsapi.Options{
			ExtensionID:         extensionID,
			LogServerAddress:    "127.0.0.1:0",
			AwsLambdaRuntimeAPI: "127.0.0.1:0",
		},
		ctx,
	)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "Err doing subscription request")
}
