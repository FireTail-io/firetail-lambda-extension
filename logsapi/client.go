package logsapi

import (
	"context"
	"firetail-lambda-extension/firetail"
	"net/http"
	"sync"

	"github.com/pkg/errors"
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
	err := options.loadEnvVars()
	if err != nil {
		return nil, err
	}

	client := &Client{
		recordsChannel:    make(chan firetail.Record, options.recordsBufferSize),
		errCallback:       options.ErrCallback,
		httpServer:        &http.Server{Addr: options.LogServerAddress},
		receiverWaitgroup: &sync.WaitGroup{},
		maxBatchSize:      options.maxBatchSize,
		batchCallback:     options.BatchCallback,
	}

	err = subscribeToLogsApi(options.awsLambdaRuntimeAPI, options.ExtensionID)
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/", client.logsApiHandler)

	client.receiverWaitgroup.Add(1)
	go client.recordReceiver()

	return client, nil
}

func (c *Client) Start(ctx context.Context) error {
	err := c.httpServer.ListenAndServe()

	if err != http.ErrServerClosed {
		err = errors.WithMessage(err, "Log server closed unexpectedly")
		c.errCallback(err)
	} else if err != nil {
		c.errCallback(err)
	}

	return err
}

func (c *Client) Shutdown(ctx context.Context) error {
	err := c.httpServer.Shutdown(ctx)
	close(c.recordsChannel)
	return err
}
