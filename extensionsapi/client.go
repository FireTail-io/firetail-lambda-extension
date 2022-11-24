package extensionsapi

import (
	"fmt"
	"net/http"
	"os"
)

// Client is a simple client for the Lambda Extensions API
type Client struct {
	extensionsApiUrl string
	httpClient       *http.Client
	ExtensionID      string
}

func NewClient() *Client {
	return &Client{
		extensionsApiUrl: fmt.Sprintf("http://%s/2020-01-01/extension", os.Getenv("AWS_LAMBDA_RUNTIME_API")),
		httpClient:       &http.Client{},
	}
}
