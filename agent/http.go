// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package agent

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"firetail-lambda-extension/logsapi"
)

// DefaultHttpListenerPort is used to set the URL where the logs will be sent by Logs API
const DefaultHttpListenerPort = "1234"

// LogsApiHttpListener is used to listen to the Logs API using HTTP
type LogsApiHttpListener struct {
	httpServer *http.Server
	logQueue   chan<- []byte
}

// NewLogsApiHttpListener returns a LogsApiHttpListener with the given log queue
func NewLogsApiHttpListener(lq chan<- []byte) (*LogsApiHttpListener, error) {
	return &LogsApiHttpListener{
		httpServer: nil,
		logQueue:   lq,
	}, nil
}

func ListenOnAddress() string {
	env_aws_local, ok := os.LookupEnv("AWS_SAM_LOCAL")
	if ok && "true" == env_aws_local {
		return ":" + DefaultHttpListenerPort
	}
	return "sandbox:" + DefaultHttpListenerPort
}

// Start initiates the server in a goroutine where the logs will be sent
func (s *LogsApiHttpListener) Start() (bool, error) {
	address := ListenOnAddress()
	s.httpServer = &http.Server{Addr: address}
	http.HandleFunc("/", s.http_handler)
	go func() {
		log.Printf("Serving agent on %s", address)
		err := s.httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Printf("Unexpected stop on Http Server: %v", err)
			s.Shutdown()
		} else {
			log.Printf("Http Server closed %v", err)
		}
	}()
	return true, nil
}

// http_handler handles the requests coming from the Logs API.
// Everytime Logs API sends logs, this function will read the logs from the response body
// and put them into a synchronous queue to be read by the main goroutine.
// Logging or printing besides the error cases below is not recommended if you have subscribed to receive extension logs.
// Otherwise, logging here will cause Logs API to send new logs for the printed lines which will create an infinite loop.
func (h *LogsApiHttpListener) http_handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %+v", err)
		return
	}

	// Puts the log message into the queue
	h.logQueue <- body
	if err != nil {
		log.Printf("Can't push logs to destination: %v", err)
	}
}

// Shutdown terminates the HTTP server listening for logs
func (s *LogsApiHttpListener) Shutdown() {
	if s.httpServer != nil {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			log.Printf("Failed to shutdown http server gracefully %s", err)
		} else {
			s.httpServer = nil
		}
	}
}

// HttpAgent has the listener that receives the logs and the logger that handles the received logs
type HttpAgent struct {
	listener *LogsApiHttpListener
}

// NewHttpAgent returns an agent to listen and handle logs coming from Logs API for HTTP
// Make sure the agent is initialized by calling Init(agentId) before subscription for the Logs API.
func NewHttpAgent(jq chan<- []byte) (*HttpAgent, error) {
	logsApiListener, err := NewLogsApiHttpListener(jq)
	if err != nil {
		return nil, err
	}

	return &HttpAgent{
		listener: logsApiListener,
	}, nil
}

// Init initializes the configuration for the Logs API and subscribes to the Logs API for HTTP
func (a HttpAgent) Init(agentID string) error {
	extensions_api_address, ok := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	if !ok {
		return errors.New("AWS_LAMBDA_RUNTIME_API is not set")
	}

	logsApiBaseUrl := fmt.Sprintf("http://%s", extensions_api_address)

	logsApiClient, err := logsapi.NewClient(logsApiBaseUrl)
	if err != nil {
		return err
	}

	_, err = a.listener.Start()
	if err != nil {
		return err
	}

	eventTypes := []logsapi.EventType{logsapi.Function}
	bufferingCfg := logsapi.BufferingCfg{
		MaxItems:  10000,
		MaxBytes:  262144,
		TimeoutMS: 25,
	}
	if err != nil {
		return err
	}
	destination := logsapi.Destination{
		Protocol:   logsapi.HttpProto,
		URI:        logsapi.URI(fmt.Sprintf("http://sandbox:%s", DefaultHttpListenerPort)),
		HttpMethod: logsapi.HttpPost,
		Encoding:   logsapi.JSON,
	}

	_, err = logsApiClient.Subscribe(eventTypes, bufferingCfg, destination, agentID)
	return err
}

// Shutdown finalizes the logging and terminates the listener
func (a *HttpAgent) Shutdown() {
	a.listener.Shutdown()
}
