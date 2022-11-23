package main

import (
	"context"
	"firetail-lambda-extension/extension"
	"log"
)

// awaitShutdown calls /event/next until a shutdown event is received, or the context is cancelled
func awaitShutdown(extensionClient *extension.Client, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, returning...")
			return
		default:
			log.Printf("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx) // This is a blocking call
			if err != nil {
				log.Panic(err.Error())
				return
			}

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				log.Println("Received shutdown event, returning...")
				return
			}
		}
	}
}
