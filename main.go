// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"firetail-lambda-extension/extensionsapi"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

func main() {
	// Configure logging
	extensionName := path.Base(os.Args[0])
	log.SetPrefix(fmt.Sprintf("[%s] ", extensionName))
	if isDebug, err := strconv.ParseBool(os.Getenv("FIRETAIL_EXTENSION_DEBUG")); err != nil || !isDebug {
		// If we're not in debug mode, we'll just send the logs to the void
		// log.SetOutput(nil)
	}

	// This context will be cancelled whenever a SIGTERM or SIGINT signal is received
	// We'll use it for our requests to the extensions API & to shutdown the log server
	ctx := getContext()

	// Create a Lambda Extensions API client & register our extension
	extensionClient := extensionsapi.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	_, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}

	// Now we know the extension ID, we can start the log server
	logServer, err := initLogServer(extensionClient.ExtensionID, ctx)
	if err != nil {
		panic(err)
	}
	defer logServer.Shutdown(ctx)

	// Create a goroutine to receive records from the log server and attempt to send them to Firetail
	recordReceiverWaitgroup := sync.WaitGroup{}
	recordReceiverWaitgroup.Add(1)
	go recordReceiver(logServer, &recordReceiverWaitgroup)
	defer recordReceiverWaitgroup.Wait()

	// awaitShutdown will block until a shutdown event is received, or the context is cancelled
	reason, err := awaitShutdown(extensionClient, ctx)
	if err != nil {
		panic(err)
	}
	log.Println("Shutting down, reason:", reason)

	// Sleep for 500ms to allow any final logs to be sent to the extension by the Lambda Logs API
	log.Printf("Sleeping for 500ms to allow final logs to be processed...")
	time.Sleep(500)
}
