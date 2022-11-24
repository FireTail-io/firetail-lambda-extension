package logsapi

import (
	"firetail-lambda-extension/firetail"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	DefaultRecordsBufferSize = 1000
	DefaultMaxBatchSize      = 100
	DefaultFiretailApiUrl    = "https://api.logging.eu-west-1.sandbox.firetail.app/logs/bulk"
)

type Options struct {
	// Configured in extension

	ExtensionID      string                        // The ID of the extension
	LogServerAddress string                        // The address that the log server should assume
	BatchCallback    func([]firetail.Record) error // A callback which will be provided batches of firetail records received from the Lambda Logs API
	ErrCallback      func(err error)               // A callback used for any errs raised when handling requests from the Lambda Logs API

	// Loaded from environment variables

	awsLambdaRuntimeAPI string // The URL of the Lambda Runtime API
	maxBatchSize        int    // The maximum size of a batch to provide to the BatchCallback
	recordsBufferSize   int    // The size of the records channel buffer
	firetailApiUrl      string // The URL of the Firetail Logging API used by the default BatchCallback
	firetailApiToken    string // The API token for the Firetail Logging API used by the default BatchCallback
}

func (o *Options) loadEnvVars() error {
	o.awsLambdaRuntimeAPI = os.Getenv("AWS_LAMBDA_RUNTIME_API")

	recordsBufferSizeStr := os.Getenv("FIRETAIL_LOG_BUFFER_SIZE")
	if recordsBufferSizeStr == "" {
		o.recordsBufferSize = DefaultRecordsBufferSize
	} else {
		bufferSize, err := strconv.Atoi(recordsBufferSizeStr)
		if err != nil {
			return errors.WithMessage(err, "FIRETAIL_LOG_BUFFER_SIZE invalid")
		}
		if bufferSize < 0 {
			return errors.Errorf("FIRETAIL_LOG_BUFFER_SIZE is %d but must be >= 0", bufferSize)
		}
		o.recordsBufferSize = bufferSize
	}

	maxBatchSizeStr := os.Getenv("FIRETAIL_MAX_BATCH_SIZE")
	if maxBatchSizeStr == "" {
		o.maxBatchSize = DefaultMaxBatchSize
	} else {
		maxBatchSize, err := strconv.Atoi(maxBatchSizeStr)
		if err != nil {
			return errors.WithMessage(err, "FIRETAIL_MAX_BATCH_SIZE invalid")
		}
		if maxBatchSize < 1 {
			return errors.Errorf("FIRETAIL_MAX_BATCH_SIZE is %d but must be >= 1", maxBatchSize)
		}
		o.maxBatchSize = maxBatchSize
	}

	firetailApiToken := os.Getenv("FIRETAIL_API_TOKEN")
	if firetailApiToken == "" {
		return errors.New("FIRETAIL_API_TOKEN not set")
	}
	o.firetailApiToken = firetailApiToken

	firetailApiUrl := os.Getenv("FIRETAIL_API_URL")
	if firetailApiUrl == "" {
		firetailApiUrl = DefaultFiretailApiUrl
	}
	o.firetailApiUrl = firetailApiUrl

	return nil
}

func (o *Options) setDefaults() {
	if o.BatchCallback == nil {
		o.BatchCallback = func(batch []firetail.Record) error {
			// Try to send the batch to Firetail
			log.Printf("Attempting to send batch of %d record(s) to Firetail...", len(batch))
			recordsSent, err := firetail.SendRecordsToSaaS(batch, o.firetailApiUrl, o.firetailApiToken)
			if err != nil {
				err = errors.WithMessage(err, fmt.Sprintf("Err sending %d record(s) to Firetail SaaS", recordsSent))
				log.Println(err.Error())
				return err
			}
			log.Printf("Successfully sent %d record(s) sent to Firetail.", recordsSent)
			return nil
		}
	}
	if o.ErrCallback == nil {
		o.ErrCallback = func(err error) {
			log.Println(err.Error())
		}
	}
}
