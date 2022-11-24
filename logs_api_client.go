package main

import (
	"context"
	"firetail-lambda-extension/logsapi"
	"net/http"

	"github.com/pkg/errors"
)

func initLogsApiClient(options logsapi.Options, ctx context.Context) (*logsapi.Client, error) {
	logServer, err := logsapi.NewClient(options)
	if err != nil {
		return nil, err
	}

	go func() {
		err := logServer.ListenAndServe()
		if err != http.ErrServerClosed {
			(*options.ErrCallback)(errors.WithMessage(err, "Log server closed unexpectedly"))
			logServer.Shutdown(ctx)
			return
		}
		(*options.ErrCallback)(err)
	}()

	return logServer, nil
}
