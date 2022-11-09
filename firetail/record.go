package firetail

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
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
	Event    events.APIGatewayProxyRequest `json:"event"`
	Response RecordResponse                `json:"response"`
}
type RecordResponse struct {
	StatusCode int64  `json:"statusCode"`
	Body       string `json:"body"`
}
