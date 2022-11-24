// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package extensionsapi

import (
	"fmt"
	"net/http"
)

// Client is a simple client for the Lambda Extensions API
type Client struct {
	extensionsApiUrl string
	httpClient       *http.Client
	ExtensionID      string
}

// NewClient returns a Lambda Extensions API client
func NewClient(awsLambdaRuntimeAPI string) *Client {
	extensionsApiUrl := fmt.Sprintf("http://%s/2020-01-01/extension", awsLambdaRuntimeAPI)
	return &Client{
		extensionsApiUrl: extensionsApiUrl,
		httpClient:       &http.Client{},
	}
}
