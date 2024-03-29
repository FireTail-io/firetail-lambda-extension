package logsapi

import (
	"bytes"
	"context"
	"firetail-lambda-extension/firetail"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message":"success"}`)
	}))
	defer testServer.Close()

	batchesReceived := [][]firetail.Record{}
	batchCallbackWaitgroup := &sync.WaitGroup{}
	batchCallbackWaitgroup.Add(1)
	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:])
	client, err := NewClient(Options{
		BatchCallback: func(batch []firetail.Record) error {
			batchesReceived = append(batchesReceived, batch)
			batchCallbackWaitgroup.Done()
			return nil
		},
	})
	require.Nil(t, err)
	require.NotNil(t, client)

	// ListenAndServe in separate routine & assert it closes correctly
	shutdownWaitgroup := sync.WaitGroup{}
	shutdownWaitgroup.Add(1)
	defer shutdownWaitgroup.Wait()
	go func() {
		err := client.Start(context.Background())
		assert.Equal(t, "http: Server closed", err.Error())
		shutdownWaitgroup.Done()
	}()

	// Mock a Lambda Logs API request
	testRequest := httptest.NewRequest(
		"POST",
		"http://"+client.httpServer.Addr,
		bytes.NewReader([]byte(`[{
			"time": "2022-11-23T10:20:39.660Z",
			"type": "function",
			"record": "firetail:log-ext:eyJldmVudCI6IHsidmVyc2lvbiI6ICIyLjAiLCAicm91dGVLZXkiOiAiR0VUIC90aW1lIiwgInJhd1BhdGgiOiAiL3RpbWUiLCAicmF3UXVlcnlTdHJpbmciOiAiIiwgImhlYWRlcnMiOiB7ImFjY2VwdCI6ICJ0ZXh0L2h0bWwsYXBwbGljYXRpb24veGh0bWwreG1sLGFwcGxpY2F0aW9uL3htbDtxPTAuOSxpbWFnZS9hdmlmLGltYWdlL3dlYnAsaW1hZ2UvYXBuZywqLyo7cT0wLjgsYXBwbGljYXRpb24vc2lnbmVkLWV4Y2hhbmdlO3Y9YjM7cT0wLjkiLCAiYWNjZXB0LWVuY29kaW5nIjogImd6aXAsIGRlZmxhdGUsIGJyIiwgImFjY2VwdC1sYW5ndWFnZSI6ICJlbi1HQixlbi1VUztxPTAuOSxlbjtxPTAuOCIsICJjYWNoZS1jb250cm9sIjogIm1heC1hZ2U9MCIsICJjb250ZW50LWxlbmd0aCI6ICIwIiwgImhvc3QiOiAiNGM2NzgzM2c2Ni5leGVjdXRlLWFwaS5ldS13ZXN0LTEuYW1hem9uYXdzLmNvbSIsICJzZWMtY2gtdWEiOiAiXCJHb29nbGUgQ2hyb21lXCI7dj1cIjEwN1wiLCBcIkNocm9taXVtXCI7dj1cIjEwN1wiLCBcIk5vdD1BP0JyYW5kXCI7dj1cIjI0XCIiLCAic2VjLWNoLXVhLW1vYmlsZSI6ICI/MCIsICJzZWMtY2gtdWEtcGxhdGZvcm0iOiAiXCJtYWNPU1wiIiwgInNlYy1mZXRjaC1kZXN0IjogImRvY3VtZW50IiwgInNlYy1mZXRjaC1tb2RlIjogIm5hdmlnYXRlIiwgInNlYy1mZXRjaC1zaXRlIjogIm5vbmUiLCAic2VjLWZldGNoLXVzZXIiOiAiPzEiLCAidXBncmFkZS1pbnNlY3VyZS1yZXF1ZXN0cyI6ICIxIiwgInVzZXItYWdlbnQiOiAiTW96aWxsYS81LjAgKE1hY2ludG9zaDsgSW50ZWwgTWFjIE9TIFggMTBfMTVfNykgQXBwbGVXZWJLaXQvNTM3LjM2IChLSFRNTCwgbGlrZSBHZWNrbykgQ2hyb21lLzEwNy4wLjAuMCBTYWZhcmkvNTM3LjM2IiwgIngtYW16bi10cmFjZS1pZCI6ICJSb290PTEtNjM3ZGYzZjctMmNmNDA2NGI3MDUxZDdjODc1ZWVkMzI1IiwgIngtZm9yd2FyZGVkLWZvciI6ICI3Ny4xNzMuMjkuMjkiLCAieC1mb3J3YXJkZWQtcG9ydCI6ICI0NDMiLCAieC1mb3J3YXJkZWQtcHJvdG8iOiAiaHR0cHMifSwgInJlcXVlc3RDb250ZXh0IjogeyJhY2NvdW50SWQiOiAiNDUzNjcxMjEwNDQ1IiwgImFwaUlkIjogIjRjNjc4MzNnNjYiLCAiZG9tYWluTmFtZSI6ICI0YzY3ODMzZzY2LmV4ZWN1dGUtYXBpLmV1LXdlc3QtMS5hbWF6b25hd3MuY29tIiwgImRvbWFpblByZWZpeCI6ICI0YzY3ODMzZzY2IiwgImh0dHAiOiB7Im1ldGhvZCI6ICJHRVQiLCAicGF0aCI6ICIvdGltZSIsICJwcm90b2NvbCI6ICJIVFRQLzEuMSIsICJzb3VyY2VJcCI6ICI3Ny4xNzMuMjkuMjkiLCAidXNlckFnZW50IjogIk1vemlsbGEvNS4wIChNYWNpbnRvc2g7IEludGVsIE1hYyBPUyBYIDEwXzE1XzcpIEFwcGxlV2ViS2l0LzUzNy4zNiAoS0hUTUwsIGxpa2UgR2Vja28pIENocm9tZS8xMDcuMC4wLjAgU2FmYXJpLzUzNy4zNiJ9LCAicmVxdWVzdElkIjogImNETU91aTd3am9FRU15QT0iLCAicm91dGVLZXkiOiAiR0VUIC90aW1lIiwgInN0YWdlIjogIiRkZWZhdWx0IiwgInRpbWUiOiAiMjMvTm92LzIwMjI6MTA6MjA6MzkgKzAwMDAiLCAidGltZUVwb2NoIjogMTY2OTE5ODgzOTYyN30sICJpc0Jhc2U2NEVuY29kZWQiOiBmYWxzZX0sICJyZXNwb25zZSI6IHsic3RhdHVzQ29kZSI6IDIwMCwgImJvZHkiOiAie1wibWVzc2FnZVwiOiBcIkhlbGxvLCB0aGUgY3VycmVudCB0aW1lIGlzIDEwOjIwOjM5LjY2MDI3MFwifSJ9LCAiZXhlY3V0aW9uX3RpbWUiOiAwLjA0ODYzNzM5MDEzNjcxODc1fQ==\n"
		}]`)),
	)
	recorder := httptest.NewRecorder()
	client.logsApiHandler(recorder, testRequest)

	// Test the logs API client gave a 200 response
	result := recorder.Result()
	assert.Equal(t, 200, result.StatusCode)

	// Wait until the batch callback has received at least 1 batch
	batchCallbackWaitgroup.Wait()

	// Test the logs API client provided the callback with exactly 1 batch containing exactly 1 record
	require.Len(t, batchesReceived, 1)
	require.Len(t, batchesReceived[0], 1)

	// Test that the record was decoded properly
	assert.Equal(t, 0.04863739013671875, batchesReceived[0][0].ExecutionTime)
	assert.Equal(t, int64(200), batchesReceived[0][0].Response.StatusCode)
	assert.Equal(t, "{\"message\": \"Hello, the current time is 10:20:39.660270\"}", batchesReceived[0][0].Response.Body)

	// Test the client shuts down successfully
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = client.Shutdown(ctx)
	assert.Nil(t, err)
}

func TestNewClientSubscribeBroken(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"message":"internal server error"}`)
	}))
	defer testServer.Close()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:])
	client, err := NewClient(Options{})
	assert.Nil(t, client)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "failed with status code 500 and response body {\"message\":\"internal server error\"}")
}

func TestNewClientSubscribeLocal(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"message":"success"}`)
	}))
	defer testServer.Close()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:])
	client, err := NewClient(Options{})
	assert.Nil(t, client)
	require.NotNil(t, err)

	assert.Contains(t, err.Error(), "Logs API not supported. Extension may be running in local sandbox.")
}

func TestNewClientSubscribeNoServer(t *testing.T) {
	t.Setenv("AWS_LAMBDA_RUNTIME_API", "127.0.0.1:0")
	client, err := NewClient(Options{})
	assert.Nil(t, client)
	require.NotNil(t, err)

	// This may vary depending upon the runtime
	assert.Contains(t, err.Error(), "Err doing subscription request: Put \"http://127.0.0.1:0/2020-08-15/logs\"")
}

func TestNewClientWithInvalidLogServerAddress(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	runtimeApiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"eventType": "SHUTDOWN"}`)
	}))
	defer runtimeApiServer.Close()

	extensionID := "TEST_EXTENSION_ID"
	ctx := context.Background()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(runtimeApiServer.URL, ":")[1:], ":")[2:])
	logsApiClient, err := NewClient(
		Options{
			ExtensionID:      extensionID,
			LogServerAddress: ":::",
		},
	)
	require.Nil(t, err)

	var closeErr error
	errWaitgroup := sync.WaitGroup{}
	errWaitgroup.Add(1)
	go func() {
		closeErr = logsApiClient.Start(ctx)
		errWaitgroup.Done()
	}()

	errWaitgroup.Wait()
	assert.Equal(t, "Log server closed unexpectedly: listen tcp: address :::: too many colons in address", closeErr.Error())
}

func TestNewClientWithNoRuntimeApi(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)

	extensionID := "TEST_EXTENSION_ID"

	t.Setenv("AWS_LAMBDA_RUNTIME_API", "127.0.0.1:0")
	_, err := NewClient(
		Options{
			ExtensionID:      extensionID,
			LogServerAddress: "127.0.0.1:0",
		},
	)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "Err doing subscription request")
}
