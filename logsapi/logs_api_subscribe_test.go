package logsapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeToLogsApiBadUrl(t *testing.T) {
	err := subscribeToLogsApi("\n", "TEST_EXTENSION_ID")
	require.NotNil(t, err)
	assert.Equal(t, "Err creating subscription request: parse \"http://\\n/2020-08-15/logs\": net/url: invalid control character in URL", err.Error())
}
