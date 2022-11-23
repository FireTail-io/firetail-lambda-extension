package logsapi

import (
	"bytes"
	"log"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetDefaults(t *testing.T) {
	testOptions := Options{}
	testOptions.setDefaults()
	assert.Equal(t, &defaultErrCallback, testOptions.ErrCallback)
	assert.Equal(t, DefaultRecordsBufferSize, testOptions.RecordsBufferSize)
}

func TestDefaultErrCallback(t *testing.T) {
	testErr := errors.New("Test Error")

	logBuffer := bytes.Buffer{}
	log.SetOutput(&logBuffer)

	defaultErrCallback(testErr)

	logOutput, err := logBuffer.ReadString('\n')
	require.Nil(t, err)

	assert.Contains(t, logOutput, testErr.Error())
}
