package extensionsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// EventType represents the type of events recieved from /event/next
type EventType string

const (
	Invoke   EventType = "INVOKE"   // a lambda invoke
	Shutdown EventType = "SHUTDOWN" // shutdown event for the environment

	extensionNameHeader      = "Lambda-Extension-Name"
	extensionIdentiferHeader = "Lambda-Extension-Identifier"
	extensionErrorType       = "Lambda-Extension-Function-Error-Type"
)

// RegisterResponse is the body of the response for /register
type RegisterResponse struct {
	FunctionName    string `json:"functionName"`
	FunctionVersion string `json:"functionVersion"`
	Handler         string `json:"handler"`
}

// Register will register the extension with the Extensions API
func (e *Client) Register(ctx context.Context, extensionName string) (*RegisterResponse, error) {
	const action = "/register"
	url := e.extensionsApiUrl + action

	reqBody, err := json.Marshal(map[string]interface{}{
		"events": []EventType{Invoke, Shutdown},
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set(extensionNameHeader, extensionName)
	httpRes, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status %s", httpRes.Status)
	}
	defer httpRes.Body.Close()
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}
	res := RegisterResponse{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	e.ExtensionID = httpRes.Header.Get(extensionIdentiferHeader)
	log.Println("Extension id:", e.ExtensionID)
	return &res, nil
}
