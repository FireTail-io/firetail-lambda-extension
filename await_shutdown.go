package main

import (
	"context"
	"firetail-lambda-extension/extension"
	"log"

	"github.com/pkg/errors"
)

// awaitShutdown calls /event/next until a shutdown event is received, or the context is cancelled.
// It returns a reason, or an error, depending upon the cause of the shutdown.
func awaitShutdown(extensionClient *extension.Client, ctx context.Context) (string, error) {
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

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				return "received shutdown event", nil
			}
		}
	}
}
