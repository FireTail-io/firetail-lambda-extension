// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	"encoding/json"
	"firetail-lambda-extension/agent"
	"firetail-lambda-extension/extension"
	"firetail-lambda-extension/firetail"
	"firetail-lambda-extension/logsapi"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

func main() {
	// Setup logging prefix & log that we've started
	extensionName := path.Base(os.Args[0])
	log.SetPrefix(fmt.Sprintf("[%s] ", extensionName))
	log.Println("Started...")

	// Get API url & API token from env vars
	firetailApiToken := os.Getenv("FIRETAIL_API_TOKEN")
	if firetailApiToken == "" {
		log.Fatal("FIRETAIL_API_TOKEN not set")
	}
	firetailApiUrl := os.Getenv("FIRETAIL_API_URL")
	if firetailApiUrl == "" {
		firetailApiUrl = "https://api.logging.eu-west-1.sandbox.firetail.app/logs/bulk"
		log.Printf("FIRETAIL_API_URL, defaulting to %s", firetailApiUrl)
	}

	// Create a context with which we'll perform all our actions & make a channel to receive
	// SIGTERM and SIGINT events & spawn a goroutine to call cancel() when we get one
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		log.Println("Received", s)
		log.Println("Exiting")
		cancel()
	}()

	// Create a Lambda Extensions API client & register our extension
	extensionClient := extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	_, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}

	// Create a channel down which the logsApiAgent will send events from the log API as []bytes
	logQueue := make(chan []byte)

	// Start a receiver routine for logQueue that'll run until logQueue is closed or a logsapi.RuntimeDone event is received
	go func() {
		for {
			select {
			case logBytes, open := <-logQueue:
				if !open {
					log.Println("queue channel closed, readFromLogsQueue exiting...")
					return
				}

				// Unmarshal the bytes into a LogMessages
				var logMessages logsapi.LogMessages
				err := json.Unmarshal([]byte(logBytes), &logMessages)
				if err != nil {
					log.Println("Err unmarshalling logBytes into logsapi.LogMessages, err:", err.Error())
				}

				// Extract any firetail records from the log messages
				firetailRecords, errs := firetail.ExtractFiretailRecords(logMessages)
				// Log any errs, but still continue as it's possible not all failed
				if errs != nil {
					log.Println("Errs extracting firetail records:", errs.Error())
				}
				// If there's no firetail records, then all failed or there were none, so there's nothing to do
				if len(firetailRecords) == 0 {
					log.Println("No firetail records extracted. Continuing...")
					continue
				}

				// Send the Firetail records to Firetail SaaS
				recordsSent, err := firetail.SendRecordsToSaaS(firetailRecords, firetailApiUrl, firetailApiToken)
				log.Printf("Sent %d record(s) to Firetail.\n", recordsSent)
				if err != nil {
					log.Println("Err sending record(s) to Firetail SaaS, err:", err.Error())
				}

				// Check if logMessages contains a message of type logsapi.RuntimeDone - if it does, this routine needs to exit.
				for _, logMessage := range logMessages {
					if logMessage.Type == string(logsapi.RuntimeDone) {
						log.Println("Found log message of type logsapi.RuntimeDone, logQueue receiver routine exiting...")
						return
					}
				}
			default:
				time.Sleep(time.Nanosecond)
			}
		}
	}()

	// Create a Logs API agent
	logsApiAgent, err := agent.NewHttpAgent(logQueue)
	if err != nil {
		log.Println(err)
	}

	// Subscribe to the logs API. Logs start being delivered only after the subscription happens.
	agentID := extensionClient.ExtensionID
	err = logsApiAgent.Init(agentID)
	if err != nil {
		log.Println(err)
	}

	// This for loop will block until invoke or shutdown event is received or cancelled via the context
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Println(" Waiting for event...")
			res, err := extensionClient.NextEvent(ctx) // This is a blocking call
			if err != nil {
				log.Println("Error:", err)
				log.Println("Exiting")
				return
			}

			switch res.EventType {
			case extension.Shutdown:
				// Exit if we receive a SHUTDOWN event
				logsApiAgent.Shutdown()
				close(logQueue)
				return
			}
		}
	}
}
