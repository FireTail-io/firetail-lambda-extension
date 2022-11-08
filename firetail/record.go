// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    record, err := UnmarshalRecord(bytes)
//    bytes, err = record.Marshal()

package firetail

import "encoding/json"

func UnmarshalRecord(data []byte) (Record, error) {
	var r Record
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Record) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Record struct {
	Event    RecordEvent    `json:"event"`
	Response RecordResponse `json:"response"`
}

type RecordEvent struct {
	Version         string               `json:"version"`
	RouteKey        string               `json:"routeKey"`
	RawPath         string               `json:"rawPath"`
	RawQueryString  string               `json:"rawQueryString"`
	Headers         RecordHeaders        `json:"headers"`
	RequestContext  RecordRequestContext `json:"requestContext"`
	IsBase64Encoded bool                 `json:"isBase64Encoded"`
}

type RecordHeaders struct {
	Accept          string `json:"accept"`
	ContentLength   string `json:"content-length"`
	Host            string `json:"host"`
	UserAgent       string `json:"user-agent"`
	XAmznTraceID    string `json:"x-amzn-trace-id"`
	XForwardedFor   string `json:"x-forwarded-for"`
	XForwardedPort  string `json:"x-forwarded-port"`
	XForwardedProto string `json:"x-forwarded-proto"`
}

type RecordRequestContext struct {
	AccountID    string     `json:"accountId"`
	APIID        string     `json:"apiId"`
	DomainName   string     `json:"domainName"`
	DomainPrefix string     `json:"domainPrefix"`
	HTTP         RecordHTTP `json:"http"`
	RequestID    string     `json:"requestId"`
	RouteKey     string     `json:"routeKey"`
	Stage        string     `json:"stage"`
	Time         string     `json:"time"`
	TimeEpoch    int64      `json:"timeEpoch"`
}

type RecordHTTP struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Protocol  string `json:"protocol"`
	SourceIP  string `json:"sourceIp"`
	UserAgent string `json:"userAgent"`
}

type RecordResponse struct {
	StatusCode int64  `json:"statusCode"`
	Body       string `json:"body"`
}
