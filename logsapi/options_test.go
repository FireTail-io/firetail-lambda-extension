package logsapi

import (
	"bytes"
	"firetail-lambda-extension/firetail"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDefaults(t *testing.T) {
	testOptions := Options{}
	testOptions.setDefaults()
	assert.NotNil(t, testOptions.BatchCallback)
	assert.NotNil(t, testOptions.ErrCallback)
}

func TestLoadEnvVars(t *testing.T) {
	t.Setenv("FIRETAIL_LOG_BUFFER_SIZE", "3142")
	t.Setenv("FIRETAIL_MAX_BATCH_SIZE", "2718")
	t.Setenv("FIRETAIL_API_TOKEN", "TEST_API_TOKEN")
	t.Setenv("FIRETAIL_API_URL", "TEST_API_URL")
	testOptions := Options{}
	err := testOptions.loadEnvVars()
	require.Nil(t, err)
	assert.Equal(t, 3142, testOptions.recordsBufferSize)
	assert.Equal(t, 2718, testOptions.maxBatchSize)
	assert.Equal(t, "TEST_API_TOKEN", testOptions.firetailApiToken)
	assert.Equal(t, "TEST_API_URL", testOptions.firetailApiUrl)
}

func TestLoadEnvVarsNonIntegers(t *testing.T) {
	integerEnvVars := []string{"FIRETAIL_MAX_BATCH_SIZE", "FIRETAIL_LOG_BUFFER_SIZE"}
	t.Setenv("FIRETAIL_API_TOKEN", "TEST_TOKEN")
	for _, integerEnvVar := range integerEnvVars {
		t.Setenv(integerEnvVar, "NOT_A_NUMBER")
		testOptions := Options{}
		err := testOptions.loadEnvVars()
		require.NotNil(t, err)
		assert.Equal(
			t,
			fmt.Sprintf(
				"%s invalid: strconv.Atoi: parsing \"NOT_A_NUMBER\": invalid syntax",
				integerEnvVar,
			),
			err.Error(),
		)
	}
}

func TestLoadEnvVarsNegativeIntegers(t *testing.T) {
	integerEnvVars := map[string]int{
		"FIRETAIL_MAX_BATCH_SIZE":  1,
		"FIRETAIL_LOG_BUFFER_SIZE": 0,
	}
	t.Setenv("FIRETAIL_API_TOKEN", "TEST_TOKEN")
	for integerEnvVar, minValue := range integerEnvVars {
		t.Setenv(integerEnvVar, "-3142")
		testOptions := Options{}
		err := testOptions.loadEnvVars()
		require.NotNil(t, err)
		assert.Equal(
			t,
			fmt.Sprintf(
				"%s is -3142 but must be >= %d",
				integerEnvVar, minValue,
			),
			err.Error(),
		)
		t.Setenv(integerEnvVar, "3142")
	}
}

func TestLoadEnvVarsDefaults(t *testing.T) {
	testOptions := Options{}
	err := testOptions.loadEnvVars()
	require.Nil(t, err)
	assert.Equal(t, "", testOptions.firetailApiToken)
	assert.Equal(t, DefaultRecordsBufferSize, testOptions.recordsBufferSize)
	assert.Equal(t, DefaultMaxBatchSize, testOptions.maxBatchSize)
	assert.Equal(t, DefaultFiretailApiUrl, testOptions.firetailApiUrl)
}

func TestDefaultErrCallback(t *testing.T) {
	testOptions := Options{}
	testOptions.setDefaults()

	testErr := errors.New("Test Error")

	logBuffer := bytes.Buffer{}
	log.SetOutput(&logBuffer)

	testOptions.ErrCallback(testErr)

	logOutput, err := logBuffer.ReadString('\n')
	require.Nil(t, err)

	assert.Contains(t, logOutput, testErr.Error())
}

func TestDefaultBatchCallback(t *testing.T) {
	var requestBody []byte

	http.DefaultServeMux = new(http.ServeMux)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		requestBody, err = ioutil.ReadAll(r.Body)
		require.Nil(t, err)
		fmt.Fprintf(w, `{"message":"success"}`)
	}))
	defer testServer.Close()

	testOptions := &Options{
		firetailApiToken: "TEST_TOKEN",
		firetailApiUrl:   testServer.URL,
	}
	testOptions.setDefaults()

	err := testOptions.BatchCallback([]firetail.Record{
		{
			Event: []byte(`{"description":"test event"}`),
			Response: firetail.RecordResponse{
				StatusCode: 200,
				Body:       `{"description":"test response body"}`,
				Headers: map[string]string{
					"Test-Header-Name": "Test-Header-Value",
				},
			},
			ExecutionTime: 3.142,
		},
	})
	require.Nil(t, err)

	assert.Equal(t,
		"{\"dateCreated\":0,\"executionTime\":3.142,\"request\":{\"body\":\"\",\"headers\":{},\"httpProtocol\":\"\",\"ip\":\"\",\"method\":\"\",\"uri\":\"https://\",\"resource\":\"\"},\"response\":{\"body\":\"{\\\"description\\\":\\\"test response body\\\"}\",\"headers\":{\"Test-Header-Name\":[\"Test-Header-Value\"]},\"statusCode\":200},\"version\":\"1.0.0-alpha\",\"metadata\":{\"source\":\"lambda-extension\"}}\n",
		string(requestBody),
	)
}

func TestDefaultBatchCallbackFail(t *testing.T) {
	testOptions := &Options{
		firetailApiToken: "TEST_TOKEN",
		firetailApiUrl:   "\n",
	}
	testOptions.setDefaults()

	err := testOptions.BatchCallback([]firetail.Record{
		{
			Event: []byte(`{"description":"test event"}`),
			Response: firetail.RecordResponse{
				StatusCode: 200,
				Body:       `{"description":"test response body"}`,
				Headers: map[string]string{
					"Test-Header-Name": "Test-Header-Value",
				},
			},
			ExecutionTime: 3.142,
		},
	})
	require.NotNil(t, err)
	assert.Equal(t, "Err sending 0 record(s) to Firetail SaaS: 1 error occurred:\n\t* parse \"\\n\": net/url: invalid control character in URL\n\n", err.Error())
}
