package logsapi

import "encoding/json"

// LogMessages matches the request body used by the logs API to provide log messages to the extension
type LogMessages []LogMessage

// LogMessage matches a single log message in the LogMessages provided by the logs API in its request bodies
type LogMessage struct {
	Time   string          `json:"time"`
	Type   string          `json:"type"`
	Record json.RawMessage `json:"record"`
}
