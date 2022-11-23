package main

import (
	"context"
	"firetail-lambda-extension/extensionsapi"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMockExtensionsApiServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"eventType": "SHUTDOWN"}`)
	}))
}

func TestAwaitShutdownContextCancelled(t *testing.T) {
	extensionClient := extensionsapi.NewClient("")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()

	reason, err := awaitShutdown(extensionClient, ctx)

	assert.Nil(t, err)
	assert.Equal(t, "context cancelled", reason)
}

func TestAwaitShutdownNextEventErrs(t *testing.T) {
	extensionClient := extensionsapi.NewClient("")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	reason, err := awaitShutdown(extensionClient, ctx)

	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to get next event")
	assert.Equal(t, "", reason)

	cancel()
}

func TestAwaitShutdownShutdownEvent(t *testing.T) {
	mockExtensionsApi := getMockExtensionsApiServer()
	defer mockExtensionsApi.Close()

	extensionClient := extensionsapi.NewClient(
		strings.Join(strings.Split(mockExtensionsApi.URL, ":")[1:], ":")[2:],
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reason, err := awaitShutdown(extensionClient, ctx)
	require.Nil(t, err)
	assert.Equal(t, "received shutdown event", reason)
}
