package firetail

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

func SendRecordsToSaaS(records []Record, apiUrl, apiKey string) error {
	reqBytes := []byte{}

	for _, record := range records {
		logEntryBytes, err := json.Marshal(LogEntry{
			DateCreated:   time.Now().UnixMilli(),
			ExecutionTime: 0,
			Request: LogEntryRequest{
				Body:         record.Event.Body,
				Headers:      map[string][]string{},
				HTTPProtocol: LogEntryHTTPProtocol(record.Event.RequestContext.Protocol),
				IP:           record.Event.RequestContext.Identity.SourceIP,
				Method:       LogEntryMethod(record.Event.RequestContext.HTTPMethod),
				URI:          "https://" + record.Event.RequestContext.DomainName + record.Event.RequestContext.Path,
				Resource:     record.Event.Resource,
			},
			Response: LogEntryResponse{
				Body:       record.Response.Body,
				Headers:    map[string][]string{},
				StatusCode: record.Response.StatusCode,
			},
			Version: The100Alpha,
		})
		if err != nil {
			log.Println("Err marshalling record to bytes, err:", err.Error())
			continue
		}
		reqBytes = append(reqBytes, logEntryBytes...)
		reqBytes = append(reqBytes, '\n')
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}

	req.Header.Set("x-ft-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	if res["message"] != "success" {
		return errors.New(fmt.Sprintf("got err response from firetail api: %v, req body:\n'%s'\n", res, string(reqBytes)))
	} else {
		log.Println("Successfully sent entries to Firetail, response:", res)
	}

	return nil
}
