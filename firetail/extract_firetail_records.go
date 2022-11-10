package firetail

import (
	"encoding/base64"
	"encoding/json"
	"firetail-lambda-extension/logsapi"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
)

func ExtractFiretailRecords(logMessages logsapi.LogMessages) ([]Record, error) {
	firetailRecords := []Record{}
	var errs error

	// For each event item, if they are a function event, and their record field is a string, then try to decode
	// it as a firetail event record. If it is, append it to the slice of firetail records!
	for _, logMessage := range logMessages {

		if logMessage.Type != "function" {
			errs = multierror.Append(errs, fmt.Errorf("logMessage type is '%s', not 'function'", logMessage.Type))
			continue
		}

		var unmarshalledRecord string
		err := json.Unmarshal(logMessage.Record, &unmarshalledRecord)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Err unmarshalling event record as string, err: %s", err.Error()))
		}

		firetailRecord, err := decodeFiretailRecord(unmarshalledRecord)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Err decoding event record as firetail event, err: %s", err.Error()))
			continue
		}

		firetailRecords = append(firetailRecords, *firetailRecord)
	}

	return firetailRecords, errs
}

func decodeFiretailRecord(record string) (*Record, error) {
	recordParts := strings.Split(record, ":")

	if len(recordParts) != 3 {
		return nil, fmt.Errorf("record had %d parts when split by ':'", len(recordParts))
	}

	if recordParts[0] != "firetail" {
		return nil, fmt.Errorf("record did not have firetail prefix")
	}

	if recordParts[1] != "log-ext" {
		return nil, fmt.Errorf("firetail prefixed record did not have valid token")
	}

	recordPayload, err := base64.StdEncoding.DecodeString(recordParts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to b64 decode firetail record, err: %s", err.Error())
	}

	firetailRecord, err := UnmarshalRecord([]byte(recordPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal firetail event: %s", err.Error())
	}

	return &firetailRecord, nil
}
