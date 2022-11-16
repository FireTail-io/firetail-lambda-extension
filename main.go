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
	"strconv"
	"sync"
	"syscall"
	"time"
)

var debug bool

func debugLog(format string, a ...any) {
	if debug {
		log.Printf(format, a...)
	}
}

func main() {
	// Setup logging prefix & log that we've started
	extensionName := path.Base(os.Args[0])
	log.SetPrefix(fmt.Sprintf("[%s] ", extensionName))

	// Check if we're running in DEBUG mode
	debugEnv := os.Getenv("FIRETAIL_EXTENSION_DEBUG")
	if isDebug, err := strconv.ParseBool(debugEnv); err == nil && isDebug {
		debug = true
		debugLog("Firetail extension starting in debug mode")
	}

	// Get API url & API token from env vars
	firetailApiToken := os.Getenv("FIRETAIL_API_TOKEN")
	if firetailApiToken == "" {
		log.Fatal("FIRETAIL_API_TOKEN not set")
	}
	firetailApiUrl := os.Getenv("FIRETAIL_API_URL")
	if firetailApiUrl == "" {
		firetailApiUrl = "https://api.logging.eu-west-1.sandbox.firetail.app/logs/bulk"
		debugLog("FIRETAIL_API_URL not set, defaulting to %s", firetailApiUrl)
	}

	// Create a context with which we'll perform all our requests to the extensions API
	// & make a channel to receive SIGTERM and SIGINT events & spawn a goroutine to call
	// cancel() when we get one
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		debugLog("Received signal '%s'. Exiting...", s.String())
		cancel()
	}()

	// Create a Lambda Extensions API client & register our extension
	extensionClient := extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	_, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}

	// Create a channel down which the logsApiAgent will send events from the log API as []bytes
	logQueue := make(chan []byte, 100)

	// Create a Logs API agent
	logsApiAgent, err := agent.NewHttpAgent(logQueue)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Subscribe to the logs API. Logs start being delivered only after the subscription happens.
	agentID := extensionClient.ExtensionID
	err = logsApiAgent.Init(agentID)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Start a receiver routine for logQueue that'll run until logQueue is closed
	wg := sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	go func() {
		defer wg.Done()
		for {
			select {
			case logBytes, open := <-logQueue:
				if !open {
					debugLog("Queue channel closed & empty, logQueue recevier routine exiting...")
					return
				}

				var logMessages logsapi.LogMessages
				err := json.Unmarshal([]byte(logBytes), &logMessages)
				if err != nil {
					debugLog("Err unmarshalling logBytes into logsapi.LogMessages, err: %s", err.Error())
				}

				firetailRecords, errs := firetail.ExtractFiretailRecords(logMessages)
				// Log any errs, but still continue as it's possible not all failed
				if errs != nil {
					debugLog("Errs extracting firetail records, errs: %s", errs.Error())
				}
				// If there's no firetail records, then all failed or there were none, so there's nothing more to do
				if len(firetailRecords) == 0 {
					debugLog("No firetail records extracted. Continuing...")
					continue
				}

				debugLog("Sending %d record(s) to Firetail...", len(firetailRecords))
				recordsSent, err := firetail.SendRecordsToSaaS(firetailRecords, firetailApiUrl, firetailApiToken)
				debugLog("Sent %d record(s) to Firetail.", recordsSent)
				if err != nil {
					debugLog("Err sending record(s) to Firetail SaaS, err: %s", err.Error())
				}

			default:
				time.Sleep(time.Nanosecond)
			}
		}
	}()

	// This for loop will block until an invoke or shutdown event is received, or the context is cancelled
	for {
		select {
		case <-ctx.Done():
			debugLog("Context cancelled, exiting...")
			return
		default:
			debugLog("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx) // This is a blocking call
			if err != nil {
				log.Fatal(err.Error())
				return
			}

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				debugLog("Received extension shutdown event, sleeping for 500ms to allow final logs to arrive...")
				time.Sleep(500)
				debugLog("Exiting...")
				logsApiAgent.Shutdown()
				close(logQueue)
				return
			}
		}
	}
}
