package firetail

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var APIGatewayProxyRequest []byte = []byte(`{"event":{"event":{"version":"1.0","resource":"/hi","path":"/hi","httpMethod":"GET","headers":{"Content-Length":"0","Host":"5iagptskg6.execute-api.eu-west-2.amazonaws.com","Postman-Token":"8639a798-d0e7-420a-bd98-0c5cb16c6115","User-Agent":"PostmanRuntime/7.28.4","X-Amzn-Trace-Id":"Root=1-63761e03-7bc79fb21f90dbbe66feba18","X-Forwarded-For":"37.228.214.117","X-Forwarded-Port":"443","X-Forwarded-Proto":"https","accept":"*/*","accept-encoding":"gzip, deflate, br"},"multiValueHeaders":{"Content-Length":["0"],"Host":["5iagptskg6.execute-api.eu-west-2.amazonaws.com"],"Postman-Token":["8639a798-d0e7-420a-bd98-0c5cb16c6115"],"User-Agent":["PostmanRuntime/7.28.4"],"X-Amzn-Trace-Id":["Root=1-63761e03-7bc79fb21f90dbbe66feba18"],"X-Forwarded-For":["37.228.214.117"],"X-Forwarded-Port":["443"],"X-Forwarded-Proto":["https"],"accept":["*/*"],"accept-encoding":["gzip, deflate, br"]},"queryStringParameters":null,"multiValueQueryStringParameters":null,"requestContext":{"accountId":"453671210445","apiId":"5iagptskg6","domainName":"5iagptskg6.execute-api.eu-west-2.amazonaws.com","domainPrefix":"5iagptskg6","extendedRequestId":"bvmgijZArPEEJ0w=","httpMethod":"GET","identity":{"accessKey":null,"accountId":null,"caller":null,"cognitoAmr":null,"cognitoAuthenticationProvider":null,"cognitoAuthenticationType":null,"cognitoIdentityId":null,"cognitoIdentityPoolId":null,"principalOrgId":null,"sourceIp":"37.228.214.117","user":null,"userAgent":"PostmanRuntime/7.28.4","userArn":null},"path":"/hi","protocol":"HTTP/1.1","requestId":"bvmgijZArPEEJ0w=","requestTime":"17/Nov/2022:11:41:55 +0000","requestTimeEpoch":1668685315222,"resourceId":"GET /hi","resourcePath":"/hi","stage":"$default"},"pathParameters":null,"stageVariables":null,"body":null,"isBase64Encoded":false},"response":{"statusCode":200,"body":"{\"message\": \"Go Serverless v1.0! Your function executed successfully!\", \"input\": {\"version\": \"1.0\", \"resource\": \"/hi\", \"path\": \"/hi\", \"httpMethod\": \"GET\", \"headers\": {\"Content-Length\": \"0\", \"Host\": \"5iagptskg6.execute-api.eu-west-2.amazonaws.com\", \"Postman-Token\": \"8639a798-d0e7-420a-bd98-0c5cb16c6115\", \"User-Agent\": \"PostmanRuntime/7.28.4\", \"X-Amzn-Trace-Id\": \"Root=1-63761e03-7bc79fb21f90dbbe66feba18\", \"X-Forwarded-For\": \"37.228.214.117\", \"X-Forwarded-Port\": \"443\", \"X-Forwarded-Proto\": \"https\", \"accept\": \"*/*\", \"accept-encoding\": \"gzip, deflate, br\"}, \"multiValueHeaders\": {\"Content-Length\": [\"0\"], \"Host\": [\"5iagptskg6.execute-api.eu-west-2.amazonaws.com\"], \"Postman-Token\": [\"8639a798-d0e7-420a-bd98-0c5cb16c6115\"], \"User-Agent\": [\"PostmanRuntime/7.28.4\"], \"X-Amzn-Trace-Id\": [\"Root=1-63761e03-7bc79fb21f90dbbe66feba18\"], \"X-Forwarded-For\": [\"37.228.214.117\"], \"X-Forwarded-Port\": [\"443\"], \"X-Forwarded-Proto\": [\"https\"], \"accept\": [\"*/*\"], \"accept-encoding\": [\"gzip, deflate, br\"]}, \"queryStringParameters\": null, \"multiValueQueryStringParameters\": null, \"requestContext\": {\"accountId\": \"453671210445\", \"apiId\": \"5iagptskg6\", \"domainName\": \"5iagptskg6.execute-api.eu-west-2.amazonaws.com\", \"domainPrefix\": \"5iagptskg6\", \"extendedRequestId\": \"bvmgijZArPEEJ0w=\", \"httpMethod\": \"GET\", \"identity\": {\"accessKey\": null, \"accountId\": null, \"caller\": null, \"cognitoAmr\": null, \"cognitoAuthenticationProvider\": null, \"cognitoAuthenticationType\": null, \"cognitoIdentityId\": null, \"cognitoIdentityPoolId\": null, \"principalOrgId\": null, \"sourceIp\": \"37.228.214.117\", \"user\": null, \"userAgent\": \"PostmanRuntime/7.28.4\", \"userArn\": null}, \"path\": \"/hi\", \"protocol\": \"HTTP/1.1\", \"requestId\": \"bvmgijZArPEEJ0w=\", \"requestTime\": \"17/Nov/2022:11:41:55 +0000\", \"requestTimeEpoch\": 1668685315222, \"resourceId\": \"GET /hi\", \"resourcePath\": \"/hi\", \"stage\": \"$default\"}, \"pathParameters\": null, \"stageVariables\": null, \"body\": null, \"isBase64Encoded\": false}}"}},"response":{"statusCode":200,"body":"{\"Description\":\"This is a test response body\"}"},"execution_time":50}`)

func TestEncodeAndDecodeRecord(t *testing.T) {
	testRecord := Record{
		Event: json.RawMessage(APIGatewayProxyRequest),
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
