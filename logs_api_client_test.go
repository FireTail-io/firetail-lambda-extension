package main

import (
	"context"
	"firetail-lambda-extension/logsapi"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func getMockRuntimeApiServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"eventType": "SHUTDOWN"}`)
	}))
}

func TestInitLogsApiClient(t *testing.T) {
	runtimeApiServer := getMockRuntimeApiServer()
	defer runtimeApiServer.Close()

	extensionID := "TEST_EXTENSION_ID"
	ctx := context.Background()

	logsApiClient, err := initLogsApiClient(
		logsapi.Options{
			ExtensionID:         extensionID,
			LogServerAddress:    "127.0.0.1:0",
			AwsLambdaRuntimeAPI: strings.Join(strings.Split(runtimeApiServer.URL, ":")[1:], ":")[2:],
		},
		ctx,
	)
	require.Nil(t, err)

	err = logsApiClient.Shutdown(ctx)
	require.Nil(t, err)
}
