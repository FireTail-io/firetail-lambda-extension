package main

import (
	"bytes"
	"context"
	"firetail-lambda-extension/logsapi"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordReceiver(t *testing.T) {
	logBuffer := bytes.Buffer{}
	log.SetOutput(&logBuffer)

	t.Setenv("FIRETAIL_API_TOKEN", "TEST_TOKEN")
	t.Setenv("FIRETAIL_API_URL", "TEST_API_URL")

	testLogsApiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"success"}`)
	}))
	defer testLogsApiServer.Close()

	testClient, err := logsapi.NewClient(logsapi.Options{
		ExtensionID:         "TEST_EXTENSION_ID",
		LogServerAddress:    "127.0.0.1:0",
		AwsLambdaRuntimeAPI: strings.Join(strings.Split(testLogsApiServer.URL, ":")[1:], ":")[2:],
	})
	require.Nil(t, err)

	recordReceiverWaitgroup := &sync.WaitGroup{}
	recordReceiverWaitgroup.Add(1)
	go recordReceiver(testClient, recordReceiverWaitgroup)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = testClient.Shutdown(ctx)
	require.Nil(t, err)

	recordReceiverWaitgroup.Wait()

	logLine, err := logBuffer.ReadString('\n')
	assert.Contains(t, logLine, "Starting record receiver routine...")
	logLine, err = logBuffer.ReadString('\n')
	assert.Contains(t, logLine, "No records left to receive, exiting...")
}
