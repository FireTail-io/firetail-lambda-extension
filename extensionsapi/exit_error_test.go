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
	require.Nil(t, err)

	assert.Equal(t, "test", res.Status)
}
