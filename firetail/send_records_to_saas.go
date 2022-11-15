package firetail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-multierror"
)

// SendMessagesToSaaS takes an array of Firetail log records, and an API URL & key, and sends those records to the API provided. It returns
// an integer value representing the number of records that were included in the request sent to Firetail, and an error containing any errors
// that were encountered when Marshalling the records and attempting to send them to Firetail - note, it is possible for some logs to be sent
// to Firetail, while other logs fail due to errs while attempting to Marshal them.
func SendRecordsToSaaS(records []Record, apiUrl, apiKey string) (int, error) {
	reqBytes := []byte{}
	marshalledRecords := 0

	var errs error
	for _, record := range records {
		logEntryRequest, err := record.getLogEntryRequest()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Err creating log entry request value, err: %s", err.Error()))
			continue
		}

		logEntryBytes, err := json.Marshal(LogEntry{
			DateCreated:   time.Now().UnixMilli(),
			ExecutionTime: record.ExecutionTime,
			Request:       *logEntryRequest,
			Response: LogEntryResponse{
				Body:       record.Response.Body,
				Headers:    map[string][]string{},
				StatusCode: record.Response.StatusCode,
			},
			Version: The100Alpha,
		})
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Err marshalling record to bytes, err: %s", err.Error()))
			continue
		}

		reqBytes = append(reqBytes, logEntryBytes...)
		reqBytes = append(reqBytes, '\n')
		marshalledRecords += 1
	}

	// If there's no request bytes, there's no point making a request to Firetail
	if len(reqBytes) == 0 {
		return 0, errs
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return 0, multierror.Append(errs, err)
	}

	req.Header.Set("x-ft-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// The request has now been made, so we can now say that the marshalledRecords were included in the request sent to Firetail
		return marshalledRecords, multierror.Append(errs, err)
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	if res["message"] != "success" {
		return marshalledRecords, multierror.Append(errs, fmt.Errorf("got err response from firetail api: %v, req body:\n'%s'\n", res, string(reqBytes)))
	}

	return marshalledRecords, errs
}
