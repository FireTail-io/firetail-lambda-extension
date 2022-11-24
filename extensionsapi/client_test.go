package extensionsapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	runtimeApi := "127.0.0.1:0"
	client := NewClient(runtimeApi)
	assert.Equal(t, "http://127.0.0.1:0/2020-01-01/extension", client.extensionsApiUrl)
	assert.Equal(t, http.Client{}, *client.httpClient)
}
