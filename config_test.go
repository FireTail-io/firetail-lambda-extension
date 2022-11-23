package main

import (
	"log"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFiretailApiConfig(t *testing.T) {
	testToken, testUrl := "TEST_TOKEN", "TEST_URL"
	t.Setenv("FIRETAIL_API_TOKEN", testToken)
	t.Setenv("FIRETAIL_API_URL", testUrl)

	firetailApiToken, firetailApiUrl := getFiretailApiConfig()

	assert.Equal(t, testToken, firetailApiToken)
	assert.Equal(t, testUrl, firetailApiUrl)
}

func TestGetFiretailApiConfigNoApiUrl(t *testing.T) {
	testToken := "TEST_TOKEN"
	t.Setenv("FIRETAIL_API_TOKEN", testToken)

	firetailApiToken, firetailApiUrl := getFiretailApiConfig()

	assert.Equal(t, testToken, firetailApiToken)
	assert.Equal(t, defaultFiretailApiUrl, firetailApiUrl)
}

func TestGetFiretailApiConfigNoToken(t *testing.T) {
	testUrl := "TEST_URL"
	t.Setenv("FIRETAIL_API_URL", testUrl)

	var firetailApiToken, firetailApiUrl string

	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
		assert.Equal(t, "", firetailApiToken)
		assert.Equal(t, "", firetailApiUrl)

	}()

	firetailApiToken, firetailApiUrl = getFiretailApiConfig()
	require.False(t, true) // This should never run
}

func TestGetLogBufferSize(t *testing.T) {
	testBufferSize := 512
	t.Setenv("FIRETAIL_LOG_BUFFER_SIZE", strconv.Itoa(testBufferSize))

	bufferSize, err := getLogBufferSize()
	require.Nil(t, err)

	assert.Equal(t, testBufferSize, bufferSize)
}

func TestGetLogBufferSizeDefault(t *testing.T) {
	bufferSize, err := getLogBufferSize()
	require.Nil(t, err)
	assert.Equal(t, 1000, bufferSize)
}

func TestGetLogBufferSizeNotInteger(t *testing.T) {
	testBufferSize := "TEST_BUFFER_SIZE"
	t.Setenv("FIRETAIL_LOG_BUFFER_SIZE", testBufferSize)

	bufferSize, err := getLogBufferSize()
	require.NotNil(t, err)

	assert.Equal(t, "FIRETAIL_LOG_BUFFER_SIZE invalid:: strconv.Atoi: parsing \"TEST_BUFFER_SIZE\": invalid syntax", err.Error())
	assert.Equal(t, 0, bufferSize)
}

func TestGetLogBufferSizeLessThanZero(t *testing.T) {
	testBufferSize := -512
	t.Setenv("FIRETAIL_LOG_BUFFER_SIZE", strconv.Itoa(testBufferSize))

	bufferSize, err := getLogBufferSize()
	require.NotNil(t, err)

	assert.Equal(t, "FIRETAIL_LOG_BUFFER_SIZE is -512 but must be >= 0", err.Error())
	assert.Equal(t, 0, bufferSize)
}
