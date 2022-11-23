package logsapi

import (
	"context"
	"firetail-lambda-extension/firetail"
	"net/http"
)

type Client struct {
	recordsChannel chan firetail.Record
	errCallback    func(error)
	httpServer     *http.Server
}

func NewClient(options *Options) (*Client, error) {
	options.setDefaults()

	client := &Client{
		recordsChannel: make(chan firetail.Record, options.RecordsBufferSize),
		errCallback:    *options.ErrCallback,
		httpServer:     &http.Server{Addr: options.LogServerAddress},
	}

	err := subscribeToLogsApi(options.AwsLambdaRuntimeAPI, options.ExtensionID)
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/", client.logsApiHandler)

	return client, nil
}

func (c *Client) ListenAndServe() error {
	return c.httpServer.ListenAndServe()
}

func (c *Client) Shutdown(ctx context.Context) error {
	close(c.recordsChannel)
	return c.httpServer.Shutdown(ctx)
}
