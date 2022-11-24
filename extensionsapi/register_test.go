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
