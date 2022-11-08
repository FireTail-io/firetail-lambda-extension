// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    logEntry, err := UnmarshalLogEntry(bytes)
//    bytes, err = logEntry.Marshal()

package firetail

import "encoding/json"

func UnmarshalLogEntry(data []byte) (LogEntry, error) {
	var r LogEntry
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *LogEntry) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// All the information required to make a logging entry in Firetail
type LogEntry struct {
	DateCreated   int64            `json:"dateCreated"`   // The time the request was logged in UNIX milliseconds
	ExecutionTime float64          `json:"executionTime"` // The time elapsed during the execution required to respond to the request, in milliseconds
	Request       LogEntryRequest  `json:"request"`
	Response      LogEntryResponse `json:"response"`
	Version       LogEntryVersion  `json:"version"` // The version of the firetail logging schema used
}

type LogEntryRequest struct {
	Body         string               `json:"body"`         // The request body, stringified
	Headers      map[string][]string  `json:"headers"`      // The request headers
	HTTPProtocol LogEntryHTTPProtocol `json:"httpProtocol"` // The HTTP protocol used in the request
	IP           string               `json:"ip"`           // The source IP of the request
	Method       LogEntryMethod       `json:"method"`       // The request method. Src for allowed values can be found here: <a; href='https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods'>https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods</a>.
	URI          string               `json:"uri"`          // The URI the request was made to
	Resource     string               `json:"resource"`     // The resource path that the request matched up to in the OpenAPI spec
}

type LogEntryResponse struct {
	Body       string              `json:"body"`    // The response body, stringified
	Headers    map[string][]string `json:"headers"` // The response headers
	StatusCode int64               `json:"statusCode"`
}

// The HTTP protocol used in the request
type LogEntryHTTPProtocol string

const (
	HTTP10 LogEntryHTTPProtocol = "HTTP/1.0"
	HTTP11 LogEntryHTTPProtocol = "HTTP/1.1"
	HTTP2  LogEntryHTTPProtocol = "HTTP/2"
	HTTP3  LogEntryHTTPProtocol = "HTTP/3"
)

// The request method. Src for allowed values can be found here: <a
// href='https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods'>https://www.iana.org/assignments/http-methods/http-methods.xhtml#methods</a>.
type LogEntryMethod string

const (
	ACL               LogEntryMethod = "ACL"
	BaselineControl   LogEntryMethod = "BASELINE-CONTROL"
	Bind              LogEntryMethod = "BIND"
	Checkin           LogEntryMethod = "CHECKIN"
	Checkout          LogEntryMethod = "CHECKOUT"
	Connect           LogEntryMethod = "CONNECT"
	Copy              LogEntryMethod = "COPY"
	Delete            LogEntryMethod = "DELETE"
	Empty             LogEntryMethod = "*"
	Get               LogEntryMethod = "GET"
	Head              LogEntryMethod = "HEAD"
	Label             LogEntryMethod = "LABEL"
	Link              LogEntryMethod = "LINK"
	Lock              LogEntryMethod = "LOCK"
	Merge             LogEntryMethod = "MERGE"
	Mkactivity        LogEntryMethod = "MKACTIVITY"
	Mkcalendar        LogEntryMethod = "MKCALENDAR"
	Mkcol             LogEntryMethod = "MKCOL"
	Mkredirectref     LogEntryMethod = "MKREDIRECTREF"
	Mkworkspace       LogEntryMethod = "MKWORKSPACE"
	Move              LogEntryMethod = "MOVE"
	Options           LogEntryMethod = "OPTIONS"
	Orderpatch        LogEntryMethod = "ORDERPATCH"
	Patch             LogEntryMethod = "PATCH"
	Post              LogEntryMethod = "POST"
	Pri               LogEntryMethod = "PRI"
	Propfind          LogEntryMethod = "PROPFIND"
	Proppatch         LogEntryMethod = "PROPPATCH"
	Put               LogEntryMethod = "PUT"
	Rebind            LogEntryMethod = "REBIND"
	Report            LogEntryMethod = "REPORT"
	Search            LogEntryMethod = "SEARCH"
	Trace             LogEntryMethod = "TRACE"
	Unbind            LogEntryMethod = "UNBIND"
	Uncheckout        LogEntryMethod = "UNCHECKOUT"
	Unlink            LogEntryMethod = "UNLINK"
	Unlock            LogEntryMethod = "UNLOCK"
	Update            LogEntryMethod = "UPDATE"
	Updateredirectref LogEntryMethod = "UPDATEREDIRECTREF"
	VersionControl    LogEntryMethod = "VERSION-CONTROL"
)

// The version of the firetail logging schema used
type LogEntryVersion string

const (
	The100Alpha LogEntryVersion = "1.0.0-alpha"
)
