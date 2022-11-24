package logsapi

import (
	"bytes"
	"fmt"
	"log"
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
	}
}

func TestLoadEnvVarsNoApiToken(t *testing.T) {
	testOptions := Options{}
	err := testOptions.loadEnvVars()
	require.NotNil(t, err)
	assert.Equal(t, "FIRETAIL_API_TOKEN not set", err.Error())
}

func TestLoadEnvVarsDefaults(t *testing.T) {
	t.Setenv("FIRETAIL_API_TOKEN", "TEST_TOKEN")
	testOptions := Options{}
	err := testOptions.loadEnvVars()
	require.Nil(t, err)
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
