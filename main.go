// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	"firetail-lambda-extension/agent"
	"firetail-lambda-extension/extension"
	"firetail-lambda-extension/firetail"
	"firetail-lambda-extension/logsapi"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Setup logging prefix & log that we've started
	extensionName := path.Base(os.Args[0])
	log.SetPrefix(fmt.Sprintf("[%s] ", extensionName))
	log.Println("Started...")

	// Get API url & API token from env vars
	firetailApiUrl := os.Getenv("FIRETAIL_API_URL")
	if firetailApiUrl == "" {
		log.Fatal("FIRETAIL_API_URL not set")
	}
	firetailApiToken := os.Getenv("FIRETAIL_API_TOKEN")
	if firetailApiToken == "" {
		log.Fatal("FIRETAIL_API_TOKEN not set")
	}
	firetailLogsUuid := os.Getenv("FIRETAIL_LOGGING_UUID")
	if firetailLogsUuid == "" {
		log.Fatal("FIRETAIL_LOGGING_UUID not set")
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
			case logs, open := <-logQueue:
				if !open {
					log.Println("queue channel closed, readFromLogsQueue exiting...")
					return
				}

				// Extract any firetail records from the log bytes
				firetailRecords, errs := firetail.ExtractFiretailRecords(logs, firetailLogsUuid)

				// Log any errs
				if len(errs) > 0 {
					errString := "'"
					for i, err := range errs {
						if i > 0 {
							errString += "', '"
						}
						errString += err.Error()
					}
					errString += "'"
					log.Println("Errs extracting firetail records:", errString)
				}

				// Send the logs to Firetail SaaS
				err := firetail.SendRecordsToSaaS(firetailRecords, firetailApiUrl, firetailApiToken)
				if err != nil {
					log.Println("Err sending logs to Firetail SaaS, err:", err.Error())
				}

				// Check if logs contains logsapi.RuntimeDone - if it does, this routine needs to exit.
				// TODO: this should be FAR more strict - if the string "platform.runtimeDone" appears in ANY function logs, this routine will exit.
				if strings.Contains(fmt.Sprintf("%v", logs), string(logsapi.RuntimeDone)) {
					log.Println("found logsapi.RuntimeDone in logs string, readFromLogsQueue exiting...")
					return
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
