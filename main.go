package main

import (
	"firetail-lambda-extension/extensionsapi"
	"firetail-lambda-extension/firetail"
	"firetail-lambda-extension/logsapi"
	"firetail-lambda-extension/proxy"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

func main() {
	// Configure logging
	extensionName := path.Base(os.Args[0])
	log.SetPrefix(fmt.Sprintf("[%s] ", extensionName))
	isDebug, err := strconv.ParseBool(os.Getenv("FIRETAIL_EXTENSION_DEBUG"))
	if err != nil || !isDebug {
		// If we're not in debug mode, we'll just send the logs to the void
		log.SetOutput(ioutil.Discard)
	} else {
		log.Println("Firetail extension starting in debug mode.")
	}

	// Log the value of AWS_LAMBDA_EXEC_WRAPPER and AWS_LAMBDA_RUNTIME_API for debugging
	if isDebug {
		log.Println("AWS_LAMBDA_EXEC_WRAPPER:", os.Getenv("AWS_LAMBDA_EXEC_WRAPPER"))
		log.Println("AWS_LAMBDA_RUNTIME_API:", os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	}

	// If the health endpoint env var is set, make a request to it to check it's reachable
	if healthEndpoint := os.Getenv("FIRETAIL_API_URL_HEALTH"); healthEndpoint != "" {
		resp, err := http.Get(healthEndpoint)
		if err != nil {
			log.Println("Error making request to health endpoint:", err)
		} else {
			defer resp.Body.Close()
			healthResponse, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error reading health endpoint response:", err)
			} else {
				log.Println("Health endpoint response:", strconv.Itoa(resp.StatusCode), string(healthResponse))
			}
		}
	}

	// This context will be cancelled whenever a SIGTERM or SIGINT signal is received
	// We'll use it for our requests to the extensions API & to shutdown the log server
	ctx := getContext()

	// Create a Lambda Extensions API client & register our extension
	extensionClient := extensionsapi.NewClient()
	_, err = extensionClient.Register(ctx, extensionName)
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
		firetailApiUrl, firetailApiUrlSet := os.LookupEnv("FIRETAIL_API_URL")
		if !firetailApiUrlSet {
			firetailApiUrl = logsapi.DefaultFiretailApiUrl
		}
		proxyServer, err := proxy.NewProxyServer()
		if err != nil {
			panic(err)
		}
		go proxyServer.ListenAndServe()
		defer proxyServer.Shutdown(ctx)
		go firetail.RecordReceiver(
			proxyServer.RecordsChannel,
			logsapi.DefaultMaxBatchSize,
			firetailApiUrl,
			os.Getenv("FIRETAIL_API_TOKEN"),
		)
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
