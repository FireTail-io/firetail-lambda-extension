package main

import (
	"context"
	"firetail-lambda-extension/logsapi"
)

func initLogsApiClient(options logsapi.Options, ctx context.Context) (*logsapi.Client, error) {
	logServer, err := logsapi.NewClient(options)
	if err != nil {
		return nil, err
	}

	return logServer, nil
}
