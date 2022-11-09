package firetail

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func ExtractFiretailRecords(logBytes []byte) ([]Record, []error) {
	firetailRecords := []Record{}
	errs := []error{}

	// Unmarshal the logBytes into an array of AWS Lambda Log API event items
	type LogEventItem struct {
		Time   string      `json:"time"`
		Type   string      `json:"type"`
		Record interface{} `json:"record"`
	}
	var logEventArray []LogEventItem
	err := json.Unmarshal([]byte(logBytes), &logEventArray)
	if err != nil {
		return firetailRecords, []error{err}
	}

	// For each event item, if they are a function event, and their record field is a string, then try to decode
	// it as a firetail event record. If it is, append it to the slice of firetail records!
	for _, logEvent := range logEventArray {

		if logEvent.Type != "function" {
			errs = append(errs, fmt.Errorf("logEvent type is '%s', not 'function'", logEvent.Type))
			continue
		}

		functionRecord, ok := logEvent.Record.(string)
		if !ok {
			errs = append(errs, fmt.Errorf("logEvent.Record could not be asserted to string"))
			continue
		}

		firetailRecord, err := decodeFiretailRecord(functionRecord)
		if err != nil {
			errs = append(errs, fmt.Errorf("Err decoding event record as firetail event, err: %s", err.Error()))
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
