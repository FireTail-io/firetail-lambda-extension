package logsapi

import (
	"context"
	"firetail-lambda-extension/firetail"
	"net/http"
	"sync"
)

type Client struct {
	recordsChannel    chan firetail.Record
	errCallback       func(error)
	httpServer        *http.Server
	receiverWaitgroup *sync.WaitGroup
	maxBatchSize      int
	batchCallback     func([]firetail.Record) error
}

func NewClient(options Options) (*Client, error) {
	options.setDefaults()
	options.loadEnvVars()

	client := &Client{
		recordsChannel:    make(chan firetail.Record, options.recordsBufferSize),
		errCallback:       options.ErrCallback,
		httpServer:        &http.Server{Addr: options.LogServerAddress},
		receiverWaitgroup: &sync.WaitGroup{},
		maxBatchSize:      options.maxBatchSize,
		batchCallback:     options.BatchCallback,
	}

	err := subscribeToLogsApi(options.awsLambdaRuntimeAPI, options.ExtensionID)
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/", client.logsApiHandler)

	client.receiverWaitgroup.Add(1)
	go client.recordReceiver()

	return client, nil
}

func (c *Client) ListenAndServe() error {
	return c.httpServer.ListenAndServe()
}

func (c *Client) Shutdown(ctx context.Context) error {
	close(c.recordsChannel)
	return c.httpServer.Shutdown(ctx)
}
