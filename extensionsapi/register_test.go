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

func TestRegister(t *testing.T) {
	testExtensionName := "TEST_EXTENSION_NAME"
	testRegisterResponse := RegisterResponse{
		FunctionName:    "TEST_FUNCTION_NAME",
		FunctionVersion: "TEST_FUNCTION_VERSION",
		Handler:         "TEST_FUNCTION_HANDLER",
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "{\"events\":[\"INVOKE\",\"SHUTDOWN\"]}", string(body))

		res, err := json.Marshal(testRegisterResponse)
		require.Nil(t, err)

		w.Write(res)
	}))
	defer testServer.Close()

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	ctx := context.Background()
	res, err := client.Register(ctx, testExtensionName)
	require.Nil(t, err)

	assert.Equal(t, res.FunctionName, testRegisterResponse.FunctionName)
	assert.Equal(t, res.FunctionVersion, testRegisterResponse.FunctionVersion)
	assert.Equal(t, res.Handler, testRegisterResponse.Handler)
}

func TestRegisterBadUrl(t *testing.T) {
	client := NewClient("\n")
	ctx := context.Background()
	res, err := client.Register(ctx, "TEST_EXTENSION_NAME")
	assert.Nil(t, res)
	require.NotNil(t, err)
	assert.Equal(t, "parse \"http://\\n/2020-01-01/extension/register\": net/url: invalid control character in URL", err.Error())
}

func TestRegisterNoServer(t *testing.T) {
	client := NewClient("127.0.0.1:0")
	ctx := context.Background()
	res, err := client.Register(ctx, "TEST_EXTENSION_NAME")
	assert.Nil(t, res)
	require.NotNil(t, err)
	assert.Equal(t, "Post \"http://127.0.0.1:0/2020-01-01/extension/register\": dial tcp 127.0.0.1:0: connect: can't assign requested address", err.Error())
}

func TestRegisterInternalServerError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "{\"events\":[\"INVOKE\",\"SHUTDOWN\"]}", string(body))

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"internal server error"}`))
	}))
	defer testServer.Close()

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	ctx := context.Background()
	res, err := client.Register(ctx, "TEST_EXTENSION_NAME")
	assert.Nil(t, res)
	require.NotNil(t, err)

	assert.Equal(t, "request failed with status 500 Internal Server Error", err.Error())
}

func TestRegisterInvalidResponse(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.Nil(t, err)

		assert.Equal(t, "{\"events\":[\"INVOKE\",\"SHUTDOWN\"]}", string(body))

		w.Write([]byte(`{"functionName":3142}`))
	}))
	defer testServer.Close()

	client := NewClient(
		strings.Join(strings.Split(testServer.URL, ":")[1:], ":")[2:],
	)

	ctx := context.Background()
	res, err := client.Register(ctx, "TEST_EXTENSION_NAME")
	assert.Nil(t, res)
	require.NotNil(t, err)

	assert.Equal(t, "json: cannot unmarshal number into Go struct field RegisterResponse.functionName of type string", err.Error())
}
