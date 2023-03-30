package firetail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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
		logEntryRequest, requestTime, err := record.getLogEntryRequest()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Err creating log entry request value, err: %s", err.Error()))
			continue
		}

		responseHeaders := map[string][]string{}
		for headerName, headerValue := range record.Response.Headers {
			responseHeaders[headerName] = []string{headerValue}
		}

		logEntryBytes, err := json.Marshal(LogEntry{
			DateCreated:   requestTime,
			ExecutionTime: record.ExecutionTime,
			Request:       *logEntryRequest,
			Response: LogEntryResponse{
				Body:       record.Response.Body,
				Headers:    responseHeaders,
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
		return 0, multierror.Append(errs, fmt.Errorf("Failed to marshal any Firetail records to bytes"))
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return 0, multierror.Append(errs, err)
	}

	req.Header.Set("x-ft-api-key", apiKey)

	// The execution of this request may be frozen at any time - we need to break this down so we know if the request
	// was successfully written - if it was, should we make a second request? It risks double reporting assuming the
	// request received a success response... 🤔
	// TODO: investigate above.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return marshalledRecords, multierror.Append(errs, fmt.Errorf("Failed to make log request, err: %s", err.Error()))
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return marshalledRecords, multierror.Append(errs, fmt.Errorf("Failed to decode logs API response: %s", err.Error()))
	}
	if res["message"] != "success" {
		return marshalledRecords, multierror.Append(errs, fmt.Errorf("Got err response from firetail api: %v", res))
	}

	return marshalledRecords, errs
}
