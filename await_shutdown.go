package main

import (
	"context"
	"firetail-lambda-extension/extensionsapi"
	"log"

	"github.com/pkg/errors"
)

// awaitShutdown calls /event/next until a shutdown event is received, or the context is cancelled.
// It returns a reason, or an error, depending upon the cause of the shutdown.
func awaitShutdown(extensionClient *extensionsapi.Client, ctx context.Context) (string, error) {
	for {
		select {
		case <-ctx.Done():
			log.Printf(", returning...")
			return "context cancelled", nil
		default:
			log.Printf("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx) // This is a blocking call
			if err != nil {
				return "", errors.WithMessage(err, "failed to get next event")
			}
			if res.EventType == extensionsapi.Shutdown {
				return "received shutdown event", nil
			}
		}
	}
}
