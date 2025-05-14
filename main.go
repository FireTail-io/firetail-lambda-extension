package main

import (
	"firetail-lambda-extension/extensionsapi"
	"firetail-lambda-extension/logsapi"
	"firetail-lambda-extension/proxy"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

func main() {
	// Configure logging
	extensionName := path.Base(os.Args[0])
	log.SetPrefix(fmt.Sprintf("[%s] ", extensionName))
	if isDebug, err := strconv.ParseBool(os.Getenv("FIRETAIL_EXTENSION_DEBUG")); err != nil || !isDebug {
		// If we're not in debug mode, we'll just send the logs to the void
		log.SetOutput(ioutil.Discard)
	} else {
		log.Println("Firetail extension starting in debug mode.")
	}

	// This context will be cancelled whenever a SIGTERM or SIGINT signal is received
	// We'll use it for our requests to the extensions API & to shutdown the log server
	ctx := getContext()

	// Create a Lambda Extensions API client & register our extension
	extensionClient := extensionsapi.NewClient()
	_, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	log.Println("Registered extension, ID:", extensionClient.ExtensionID)

	// In legacy mode, we use the logs API. Otherwise, we use the new proxy client.
	if isLegacy, err := strconv.ParseBool(os.Getenv("FIRETAIL_EXTENSION_LEGACY")); err == nil && isLegacy {
		// Create a logsApiClient, start it & remember to shut it down when we're done
		logsApiClient, err := logsapi.NewClient(logsapi.Options{
			ExtensionID:      extensionClient.ExtensionID,
			LogServerAddress: "sandbox:1234",
		})
		if err != nil {
			panic(err)
		}
		go logsApiClient.Start(ctx)
		defer logsApiClient.Shutdown(ctx)
	} else {
		proxyServer, err := proxy.NewProxyServer()
		if err != nil {
			panic(err)
		}
		go func() {
			if err := proxyServer.ListenAndServe(); err != nil {
				panic(err)
			}
		}()
		go func() {
			for {
				select {
				case event := <-proxyServer.EventsChannel:
					eventBody, err := io.ReadAll(event.Body)
					if err != nil {
						log.Println("Error reading event body:", err)
						continue
					}
					log.Println(
						"Received event",
						"\nStatus code:", event.StatusCode,
						"\nHeaders:", event.Header,
						"\nBody:", eventBody,
					)
				case lambdaResponse := <-proxyServer.LambdaResponseChannel:
					responseBody, err := io.ReadAll(lambdaResponse.Body)
					if err != nil {
						log.Println("Error reading response body:", err)
						continue
					}
					log.Println(
						"Received lambda response",
						"\nHeaders:", lambdaResponse.Header,
						"\nBody:", responseBody,
					)
				}
			}
		}()
	}

	// awaitShutdown will block until a shutdown event is received, or the context is cancelled
	reason, err := awaitShutdown(extensionClient, ctx)
	if err != nil {
		panic(err)
	}
	log.Println("Shutting down, reason:", reason)

	// Sleep for 500ms to allow any final logs to be sent to the extension by the Lambda Logs API
	log.Printf("Sleeping for 500ms to allow final logs to be processed...")
	time.Sleep(500 * time.Millisecond)
}
