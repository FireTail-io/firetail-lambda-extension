package logsapiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// bufferingCfg is the configuration set for receiving logs from Logs API. Whichever of the conditions below is met first, the logs will be sent
type bufferingCfg struct {
	MaxItems  uint32 `json:"maxItems"`  // the maximum number of events to be buffered in memory. (default: 10000, minimum: 1000, maximum: 10000)
	MaxBytes  uint32 `json:"maxBytes"`  // the maximum size in bytes of the logs to be buffered in memory. (default: 262144, minimum: 262144, maximum: 1048576)
	TimeoutMS uint32 `json:"timeoutMs"` // the maximum time (in milliseconds) for a batch to be buffered. (default: 1000, minimum: 100, maximum: 30000)
}

// destination is the configuration for listeners who would like to receive logs with HTTP
type destination struct {
	Protocol   string `json:"protocol"` // the protocol the logs will be sent with (HTTP or TCP)
	URI        string `json:"URI"`      // the endpoint where the logs will be sent to
	HttpMethod string `json:"method"`   // represents the HTTP method used to receive logs from Logs API (PUT or POST)
	Encoding   string `json:"encoding"` // denotes what the content is encoded in (XML or JSON)
}

// SubscribeRequest is the request body that is sent to Logs API on subscribe
type subscribeRequest struct {
	SchemaVersion string       `json:"schemaVersion"` // The schema version used in the logs requests (latest is 2021-03-18)
	EventTypes    []string     `json:"types"`         // An array of event types to subscribe to (platform, extension or runtime)
	BufferingCfg  bufferingCfg `json:"buffering"`
	Destination   destination  `json:"destination"`
}

// Subscribes the log server to the Lambda Logs API.
// Takes two arguments:
// 1. The URL of the Lambda Runtime API to make the subscription request to
// 2. The ID of the extension
func subscribeToLogsApi(extensionsApiUrl, extensionID string) error {
	requestBytes, err := json.Marshal(subscribeRequest{
		SchemaVersion: "2021-03-18",
		EventTypes:    []string{"function"},
		BufferingCfg: bufferingCfg{
			MaxItems:  10000,
			MaxBytes:  262144,
			TimeoutMS: 25,
		},
		Destination: destination{
			Protocol:   "HTTP",
			URI:        "http://sandbox:1234",
			HttpMethod: "POST",
			Encoding:   "JSON",
		},
	})
	if err != nil {
		return errors.WithMessage(err, "Err marshalling subscription request bytes")
	}

	subscriptionRequest, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("http://%s/2020-08-15/logs", extensionsApiUrl),
		bytes.NewReader(requestBytes),
	)
	if err != nil {
		return errors.WithMessage(err, "Err creating subscription request")
	}
	subscriptionRequest.Header.Add("Lambda-Extension-Identifier", extensionID)

	resp, err := (&http.Client{}).Do(subscriptionRequest)
	if err != nil {
		return errors.WithMessage(err, "Err doing subscription request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return errors.New("Logs API not supported. Extension may be running in local sandbox.")
	} else if resp.StatusCode != http.StatusOK {
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.WithMessage(err, "Failed to read response to subscription request")
		}
		return errors.Errorf("Subscription request to %s failed with status code %d and response body %s", extensionsApiUrl, resp.StatusCode, string(responseBody))
	}

	return nil
}
