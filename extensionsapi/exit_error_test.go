package extensionsapi

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "", string(body))

		w.Write([]byte(`{"status":"test"}`))
	}))
	defer testServer.Close()

	ctx := context.Background()
	testErrorType := "TEST_ERROR_TYPE"

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	res, err := client.ExitError(ctx, testErrorType)
	require.NotNil(t, res)
	require.Nil(t, err)

	assert.Equal(t, "test", res.Status)
}

func TestExitErrorBadUrl(t *testing.T) {
	ctx := context.Background()
	testErrorType := "TEST_ERROR_TYPE"

	client := NewClient("\n")

	res, err := client.ExitError(ctx, testErrorType)
	assert.Nil(t, res)
	require.NotNil(t, err)
	assert.Equal(t, "parse \"http://\\n/2020-01-01/extension/exit/error\": net/url: invalid control character in URL", err.Error())
}

func TestExitErrorNoServer(t *testing.T) {
	ctx := context.Background()
	testErrorType := "TEST_ERROR_TYPE"

	client := NewClient("127.0.0.1:0")

	res, err := client.ExitError(ctx, testErrorType)
	assert.Nil(t, res)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "Post \"http://127.0.0.1:0/2020-01-01/extension/exit/error\": dial tcp 127.0.0.1:0:")
}

func TestExitErrorInternalServerError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "", string(body))

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"internal server error"}`))
	}))
	defer testServer.Close()

	ctx := context.Background()
	testErrorType := "TEST_ERROR_TYPE"

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	res, err := client.ExitError(ctx, testErrorType)
	assert.Nil(t, res)
	require.NotNil(t, err)

	assert.Equal(t, "request failed with status 500 Internal Server Error", err.Error())
}

func TestExitErrorInvalidResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "", string(body))

		w.Write([]byte(`{"status":3142}`))
	}))
	defer testServer.Close()

	ctx := context.Background()
	testErrorType := "TEST_ERROR_TYPE"

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	res, err := client.ExitError(ctx, testErrorType)
	assert.Nil(t, res)
	require.NotNil(t, err)

	assert.Equal(t, "json: cannot unmarshal number into Go struct field StatusResponse.status of type string", err.Error())
}
