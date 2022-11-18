package firetail

import (
	"encoding/json"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getNewAPIGatewayProxyRequest() events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		Resource:   "/hi",
		Path:       "/hi",
		HTTPMethod: "GET",
		Headers: map[string]string{
			"Content-Length":    "0",
			"Host":              "5iagptskg6.execute-api.eu-west-2.amazonaws.com",
			"Postman-Token":     "8639a798-d0e7-420a-bd98-0c5cb16c6115",
			"User-Agent":        "PostmanRuntime/7.28.4",
			"X-Amzn-Trace-Id":   "Root=1-63761e03-7bc79fb21f90dbbe66feba18",
			"X-Forwarded-For":   "37.228.214.117",
			"X-Forwarded-Port":  "443",
			"X-Forwarded-Proto": "https",
			"accept":            "*/*",
			"accept-encoding":   "gzip, deflate, br",
		},
		MultiValueHeaders: map[string][]string{
			"Content-Length":    {"0"},
			"Host":              {"5iagptskg6.execute-api.eu-west-2.amazonaws.com"},
			"Postman-Token":     {"8639a798-d0e7-420a-bd98-0c5cb16c6115"},
			"User-Agent":        {"PostmanRuntime/7.28.4"},
			"X-Amzn-Trace-Id":   {"Root=1-63761e03-7bc79fb21f90dbbe66feba18"},
			"X-Forwarded-For":   {"37.228.214.117"},
			"X-Forwarded-Port":  {"443"},
			"X-Forwarded-Proto": {"https"},
			"accept":            {"*/*"},
			"accept-encoding":   {"gzip, deflate, br"},
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			AccountID:         "453671210445",
			APIID:             "5iagptskg6",
			DomainName:        "5iagptskg6.execute-api.eu-west-2.amazonaws.com",
			DomainPrefix:      "5iagptskg6",
			ExtendedRequestID: "bvmgijZArPEEJ0w=",
			HTTPMethod:        "GET",
			Identity: events.APIGatewayRequestIdentity{
				SourceIP:  "37.228.214.117",
				UserAgent: "PostmanRuntime/7.28.4",
			},
			Path:             "/hi",
			Protocol:         "HTTP/1.1",
			RequestID:        "bvmgijZArPEEJ0w=",
			RequestTime:      "17/Nov/2022:11:41:55 +0000",
			RequestTimeEpoch: 1668685315222,
			ResourceID:       "GET /hi",
			ResourcePath:     "/hi",
			Stage:            "$default",
		},
	}
}

func getNewAPIGatewayV2HTTPRequest() events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		Version:  "2.0",
		RouteKey: "GET /hi",
		RawPath:  "/hi",
		Headers: map[string]string{
			"accept":            "*/*",
			"accept-encoding":   "gzip, deflate, br",
			"content-length":    "0",
			"host":              "5iagptskg6.execute-api.eu-west-2.amazonaws.com",
			"postman-token":     "071909b8-8176-47cb-8b36-cd0e8ea2081c",
			"user-agent":        "PostmanRuntime/7.28.4",
			"x-amzn-trace-id":   "Root=1-63761dac-7dc96ebc6ea580e704c4a0f2",
			"x-forwarded-for":   "37.228.214.117",
			"x-forwarded-port":  "443",
			"x-forwarded-proto": "https",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			AccountID:    "453671210445",
			APIID:        "5iagptskg6",
			DomainName:   "5iagptskg6.execute-api.eu-west-2.amazonaws.com",
			DomainPrefix: "5iagptskg6",
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method:    "GET",
				Path:      "/hi",
				Protocol:  "HTTP/1.1",
				SourceIP:  "37.228.214.117",
				UserAgent: "PostmanRuntime/7.28.4",
			},
			RequestID: "bvmS7gV3LPEEJbg=",
			RouteKey:  "GET /hi",
			Stage:     "$default",
			Time:      "17/Nov/2022:11:40:28 +0000",
			TimeEpoch: 1668685228137,
		},
	}
}

func TestEncodeAndDecodeRecord(t *testing.T) {
	apiGatewayProxyRequestBytes, err := json.Marshal(getNewAPIGatewayProxyRequest())
	require.Nil(t, err)

	testRecord := Record{
		Event: json.RawMessage(apiGatewayProxyRequestBytes),
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "{\"Description\":\"This is a test response body\"}",
		},
		ExecutionTime: 50,
	}

	testRecordBytes, err := testRecord.Marshal()
	require.Nil(t, err)

	unmarshalledRecord, err := UnmarshalRecord(testRecordBytes)
	require.Nil(t, err)
	assert.Equal(t, testRecord, unmarshalledRecord)

	remarshalledRecordBytes, err := unmarshalledRecord.Marshal()
	require.Nil(t, err)
	assert.Equal(t, testRecordBytes, remarshalledRecordBytes)

	log.Println(string(remarshalledRecordBytes))
}

func TestGetLogEntryRequestAPIGatewayProxyRequest(t *testing.T) {
	apiGatewayProxyRequest := getNewAPIGatewayProxyRequest()
	apiGatewayProxyRequestBytes, err := json.Marshal(apiGatewayProxyRequest)
	require.Nil(t, err)

	testRecord := Record{
		Event: json.RawMessage(apiGatewayProxyRequestBytes),
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "{\"Description\":\"This is a test response body\"}",
		},
		ExecutionTime: 50,
	}

	logEntry, requestAt, err := testRecord.getLogEntryRequest()
	require.Nil(t, err)

	assert.Equal(t, int64(1668685315222), requestAt)
	assert.Equal(t, apiGatewayProxyRequest.Body, logEntry.Body)
	assert.Equal(t, apiGatewayProxyRequest.MultiValueHeaders, logEntry.Headers)
	assert.Equal(t, apiGatewayProxyRequest.RequestContext.Protocol, string(logEntry.HTTPProtocol))
	assert.Equal(t, apiGatewayProxyRequest.RequestContext.Identity.SourceIP, logEntry.IP)
	assert.Equal(t, apiGatewayProxyRequest.RequestContext.HTTPMethod, string(logEntry.Method))
	assert.Equal(t, "https://"+apiGatewayProxyRequest.RequestContext.DomainName+apiGatewayProxyRequest.RequestContext.Path, logEntry.URI)
	assert.Equal(t, apiGatewayProxyRequest.Resource, logEntry.Resource)
}

func TestGetLogEntryRequestAPIGatewayProxyRequestWithNoRequestHeaders(t *testing.T) {
	apiGatewayProxyRequest := getNewAPIGatewayProxyRequest()
	// Unfortunately, we can't strip out the headers by setting apiGatewayProxyRequest.Headers to nil
	// and marshalling it here, as the Headers field of the events.APIGatewayProxyRequest has not been
	// tagged as omitempty. 🥲
	apiGatewayProxyRequestBytes := []byte(`{"resource":"/hi","path":"/hi","httpMethod":"GET","queryStringParameters":null,"multiValueQueryStringParameters":null,"pathParameters":null,"stageVariables":null,"requestContext":{"accountId":"453671210445","resourceId":"GET /hi","stage":"$default","domainName":"5iagptskg6.execute-api.eu-west-2.amazonaws.com","domainPrefix":"5iagptskg6","requestId":"bvmgijZArPEEJ0w=","extendedRequestId":"bvmgijZArPEEJ0w=","protocol":"HTTP/1.1","identity":{"cognitoIdentityPoolId":"","accountId":"","cognitoIdentityId":"","caller":"","apiKey":"","apiKeyId":"","accessKey":"","sourceIp":"37.228.214.117","cognitoAuthenticationType":"","cognitoAuthenticationProvider":"","userArn":"","userAgent":"PostmanRuntime/7.28.4","user":""},"resourcePath":"/hi","path":"/hi","authorizer":null,"httpMethod":"GET","requestTime":"17/Nov/2022:11:41:55 +0000","requestTimeEpoch":1668685315222,"apiId":"5iagptskg6"},"body":""}`)

	log.Print(string(apiGatewayProxyRequestBytes))

	testRecord := Record{
		Event: json.RawMessage(apiGatewayProxyRequestBytes),
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "{\"Description\":\"This is a test response body\"}",
		},
		ExecutionTime: 50,
	}

	logEntry, requestAt, err := testRecord.getLogEntryRequest()
	require.Nil(t, err)

	assert.Equal(t, int64(1668685315222), requestAt)
	assert.Equal(t, apiGatewayProxyRequest.Body, logEntry.Body)
	assert.Equal(t, map[string][]string{}, logEntry.Headers) // Ensure it's replaced with an empty map, not nil
	assert.Equal(t, apiGatewayProxyRequest.RequestContext.Protocol, string(logEntry.HTTPProtocol))
	assert.Equal(t, apiGatewayProxyRequest.RequestContext.Identity.SourceIP, logEntry.IP)
	assert.Equal(t, apiGatewayProxyRequest.RequestContext.HTTPMethod, string(logEntry.Method))
	assert.Equal(t, "https://"+apiGatewayProxyRequest.RequestContext.DomainName+apiGatewayProxyRequest.RequestContext.Path, logEntry.URI)
	assert.Equal(t, apiGatewayProxyRequest.Resource, logEntry.Resource)
}

func TestGetLogEntryRequestAPIGatewayV2HTTPRequest(t *testing.T) {
	apiGatewayV2HTTPRequest := getNewAPIGatewayV2HTTPRequest()
	apiGatewayV2HTTPRequestBytes, err := json.Marshal(apiGatewayV2HTTPRequest)
	require.Nil(t, err)

	testRecord := Record{
		Event: json.RawMessage(apiGatewayV2HTTPRequestBytes),
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "{\"Description\":\"This is a test response body\"}",
		},
		ExecutionTime: 50,
	}

	logEntry, requestAt, err := testRecord.getLogEntryRequest()
	require.Nil(t, err)

	assert.Equal(t, int64(1668685228137), requestAt)
	assert.Equal(t, apiGatewayV2HTTPRequest.Body, logEntry.Body)
	assert.Equal(t, apiGatewayV2HTTPRequest.RequestContext.HTTP.Protocol, string(logEntry.HTTPProtocol))
	assert.Equal(t, apiGatewayV2HTTPRequest.RequestContext.HTTP.SourceIP, logEntry.IP)
	assert.Equal(t, apiGatewayV2HTTPRequest.RequestContext.HTTP.Method, string(logEntry.Method))
	assert.Equal(t, "https://"+apiGatewayV2HTTPRequest.RequestContext.DomainName+apiGatewayV2HTTPRequest.RequestContext.HTTP.Path, logEntry.URI)
	assert.Equal(t, apiGatewayV2HTTPRequest.RawPath, logEntry.Resource)

	expectedHeaders := map[string][]string{}
	for header, value := range apiGatewayV2HTTPRequest.Headers {
		expectedHeaders[header] = strings.Split(value, ",")
	}
	assert.Equal(t, expectedHeaders, logEntry.Headers)
}

func TestGetLogEntryRequestUnsupportedPayload(t *testing.T) {
	type InvalidPayload struct {
		Headers string
	}
	invalidPayload := InvalidPayload{
		Headers: "No headers here",
	}
	invalidPayloadBytes, err := json.Marshal(invalidPayload)
	require.Nil(t, err)

	testRecord := Record{
		Event: json.RawMessage(invalidPayloadBytes),
		Response: RecordResponse{
			StatusCode: 200,
			Body:       "{\"Description\":\"This is a test response body\"}",
		},
		ExecutionTime: 50,
	}

	logEntry, requestAt, err := testRecord.getLogEntryRequest()
	assert.Nil(t, logEntry)
	assert.Equal(t, int64(0), requestAt)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "json: cannot unmarshal string into Go struct field APIGatewayProxyRequest.headers of type map[string]string")
	assert.Contains(t, err.Error(), "json: cannot unmarshal string into Go struct field APIGatewayV2HTTPRequest.headers of type map[string]string")
}
