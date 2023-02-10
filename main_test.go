package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	mockExtensionsApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"eventType": "SHUTDOWN"}`)
	}))
	defer mockExtensionsApi.Close()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(mockExtensionsApi.URL, ":")[1:], ":")[2:])

	main()
}

func TestMainDebug(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	mockExtensionsApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"eventType": "SHUTDOWN"}`)
	}))
	defer mockExtensionsApi.Close()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(mockExtensionsApi.URL, ":")[1:], ":")[2:])
	t.Setenv("FIRETAIL_EXTENSION_DEBUG", "true")

	main()
}

func TestMainReturnsInNoLessThan500Milliseconds(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	mockExtensionsApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"eventType": "SHUTDOWN"}`)
	}))
	defer mockExtensionsApi.Close()

	t.Setenv("AWS_LAMBDA_RUNTIME_API", strings.Join(strings.Split(mockExtensionsApi.URL, ":")[1:], ":")[2:])

	startTime := time.Now()

	main()

	assert.Greater(t, time.Since(startTime), 500*time.Millisecond)
}
