package main

import (
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const defaultFiretailApiUrl = "https://api.logging.eu-west-1.sandbox.firetail.app/logs/bulk"

// Returns the Firetail API token and URL to use according to the env vars & defaults
func getFiretailApiConfig() (string, string) {
	firetailApiToken := os.Getenv("FIRETAIL_API_TOKEN")
	if firetailApiToken == "" {
		log.Panic("FIRETAIL_API_TOKEN not set")
	}
	firetailApiUrl := os.Getenv("FIRETAIL_API_URL")
	if firetailApiUrl == "" {
		firetailApiUrl = defaultFiretailApiUrl
		log.Printf("FIRETAIL_API_URL not set, defaulting to %s", firetailApiUrl)
	}
	return firetailApiToken, firetailApiUrl
}

// Returns the log buffer size to use according to the env var FIRETAIL_LOG_BUFFER_SIZE & defaults
func getLogBufferSize() (int, error) {
	bufferSizeStr := os.Getenv("FIRETAIL_LOG_BUFFER_SIZE")
	if bufferSizeStr == "" {
		log.Printf("FIRETAIL_LOG_BUFFER_SIZE not set")
		return 0, nil
	}

	bufferSize, err := strconv.Atoi(bufferSizeStr)
	if err != nil {
		return 0, errors.WithMessage(err, "FIRETAIL_LOG_BUFFER_SIZE invalid:")
	}
	if bufferSize < 1 {
		return 0, errors.Errorf("FIRETAIL_LOG_BUFFER_SIZE is %d but must be >= 0", bufferSize)
	}
	return bufferSize, nil
}

// Returns the URL of the Lambda Runtime API
func getRuntimeApiUrl() string {
	return os.Getenv("AWS_LAMBDA_RUNTIME_API")
}
