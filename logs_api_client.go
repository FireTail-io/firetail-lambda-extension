package main

import (
	"context"
	"firetail-lambda-extension/logsapi"
	"log"
	"net/http"
)

func initLogsApiClient(options logsapi.Options, ctx context.Context) (*logsapi.Client, error) {
	logServer, err := logsapi.NewClient(options)
	if err != nil {
		return nil, err
	}

	go func() {
		err := logServer.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Printf("Log server closed unexpectedly: %s", err.Error())
			logServer.Shutdown(ctx)
			return
		}
		log.Printf("Log server closed %s", err.Error())
	}()

	return logServer, nil
}
