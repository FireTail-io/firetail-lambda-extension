package logsapiclient

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

// logMessage matches a single log message provided by the Lambda Logs API
type logMessage struct {
	Time   string          `json:"time"`
	Type   string          `json:"type"`
	Record json.RawMessage `json:"record"`
}

func (c *Client) logsApiHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		c.errCallback(errors.WithMessage(err, "Error reading body:"))
		return
	}

	newFiretailRecords, errs := extractFiretailRecords(body)
	if errs != nil {
		c.errCallback(errs)
	}
	log.Printf("Received %d firetail records from Lambda Logs API", len(newFiretailRecords))

	for _, firetailRecord := range newFiretailRecords {
		c.recordsChannel <- firetailRecord
	}
}
