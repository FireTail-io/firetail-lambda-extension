// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    firetailRecord, err := UnmarshalFiretailRecord(bytes)
//    bytes, err = firetailRecord.Marshal()

package main

import "encoding/json"

func UnmarshalFiretailRecord(data []byte) (FiretailRecord, error) {
	var r FiretailRecord
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *FiretailRecord) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type FiretailRecord struct {
	Event    Event    `json:"event"`
	Response Response `json:"response"`
}

type Event struct {
	Version         string         `json:"version"`
	RouteKey        string         `json:"routeKey"`
	RawPath         string         `json:"rawPath"`
	RawQueryString  string         `json:"rawQueryString"`
	Headers         Headers        `json:"headers"`
	RequestContext  RequestContext `json:"requestContext"`
	IsBase64Encoded bool           `json:"isBase64Encoded"`
}

type Headers struct {
	Accept          string `json:"accept"`
	ContentLength   string `json:"content-length"`
	Host            string `json:"host"`
	UserAgent       string `json:"user-agent"`
	XAmznTraceID    string `json:"x-amzn-trace-id"`
	XForwardedFor   string `json:"x-forwarded-for"`
	XForwardedPort  string `json:"x-forwarded-port"`
	XForwardedProto string `json:"x-forwarded-proto"`
}

type RequestContext struct {
	AccountID    string `json:"accountId"`
	APIID        string `json:"apiId"`
	DomainName   string `json:"domainName"`
	DomainPrefix string `json:"domainPrefix"`
	HTTP         HTTP   `json:"http"`
	RequestID    string `json:"requestId"`
	RouteKey     string `json:"routeKey"`
	Stage        string `json:"stage"`
	Time         string `json:"time"`
	TimeEpoch    int64  `json:"timeEpoch"`
}

type HTTP struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Protocol  string `json:"protocol"`
	SourceIP  string `json:"sourceIp"`
	UserAgent string `json:"userAgent"`
}

type Response struct {
	StatusCode int64  `json:"statusCode"`
	Body       string `json:"body"`
}
