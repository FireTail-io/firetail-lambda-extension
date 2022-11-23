package main

import (
	"context"
	"firetail-lambda-extension/logsapi"
	"log"
	"net/http"
	"os"
)

func initLogServer(extensionID string, ctx context.Context) (*logsapi.Client, error) {
	logBufferSize, err := getLogBufferSize()
	if err != nil {
		return nil, err
	}

	logServer, err := logsapi.NewClient(&logsapi.Options{
		ExtensionID:         extensionID,
		RecordsBufferSize:   logBufferSize,
		LogServerAddress:    "sandbox:1234",
		AwsLambdaRuntimeAPI: os.Getenv("AWS_LAMBDA_RUNTIME_API"),
	})
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
