package firetail

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/hashicorp/go-multierror"
)

func UnmarshalRecord(data []byte) (Record, error) {
	var r Record
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Record) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Record struct {
	Event    json.RawMessage `json:"event"`
	Response RecordResponse  `json:"response"`
}
type RecordResponse struct {
	StatusCode int64  `json:"statusCode"`
	Body       string `json:"body"`
}

func (r *Record) getLogEntryRequest() (*LogEntryRequest, error) {
	var err error

	var apiGatewayV1Request events.APIGatewayProxyRequest
	apiGatewayV1RequestErr := json.Unmarshal(r.Event, &apiGatewayV1Request)
	if apiGatewayV1RequestErr == nil {
		logEntryRequest := &LogEntryRequest{
			Body:         apiGatewayV1Request.Body,
			Headers:      apiGatewayV1Request.MultiValueHeaders,
			HTTPProtocol: LogEntryHTTPProtocol(apiGatewayV1Request.RequestContext.Protocol),
			IP:           apiGatewayV1Request.RequestContext.Identity.SourceIP,
			Method:       LogEntryMethod(apiGatewayV1Request.RequestContext.HTTPMethod),
			URI:          "https://" + apiGatewayV1Request.RequestContext.DomainName + apiGatewayV1Request.RequestContext.Path,
			Resource:     apiGatewayV1Request.Resource,
		}
		for header, value := range apiGatewayV1Request.Headers {
			_, hasValues := logEntryRequest.Headers[header]
			if hasValues {
				logEntryRequest.Headers[header] = append(logEntryRequest.Headers[header], value)
			} else {
				logEntryRequest.Headers[header] = []string{value}
			}
		}
		return logEntryRequest, nil
	}
	err = multierror.Append(err, apiGatewayV1RequestErr)

	var apiGatewayV2Request events.APIGatewayV2HTTPRequest
	apiGatewayV2RequestErr := json.Unmarshal(r.Event, &apiGatewayV2Request)
	if apiGatewayV2RequestErr == nil {
		logEntryRequest := &LogEntryRequest{
			Body:         apiGatewayV2Request.Body,
			Headers:      map[string][]string{},
			HTTPProtocol: LogEntryHTTPProtocol(apiGatewayV2Request.RequestContext.HTTP.Protocol),
			IP:           apiGatewayV2Request.RequestContext.HTTP.SourceIP,
			Method:       LogEntryMethod(apiGatewayV2Request.RequestContext.HTTP.Method),
			URI:          "https://" + apiGatewayV2Request.RequestContext.DomainName + apiGatewayV2Request.RequestContext.HTTP.Path,
			Resource:     apiGatewayV2Request.RawPath,
		}
		for header, value := range apiGatewayV2Request.Headers {
			logEntryRequest.Headers[header] = []string{value}
		}
		return logEntryRequest, nil
	}
	err = multierror.Append(err, apiGatewayV2RequestErr)

	return nil, err
}
