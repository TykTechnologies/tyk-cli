package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verifies: SYS-REQ-015
func TestErrorResponse_Error(t *testing.T) {
	t.Run("returns message verbatim", func(t *testing.T) {
		e := &ErrorResponse{Status: 404, Message: "not found"}
		assert.Equal(t, "not found", e.Error())
	})

	t.Run("returns empty string when message is empty", func(t *testing.T) {
		e := &ErrorResponse{Status: 500}
		assert.Equal(t, "", e.Error())
	})

	t.Run("ignores Code and Details when building Error string", func(t *testing.T) {
		e := &ErrorResponse{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "validation failed",
			Details: map[string]interface{}{"field": "name"},
		}
		assert.Equal(t, "validation failed", e.Error())
	})
}

// Verifies: SYS-REQ-015
func TestErrorResponse_JSONRoundTrip(t *testing.T) {
	original := ErrorResponse{
		Status:  422,
		Code:    "INVALID_INPUT",
		Message: "bad payload",
		Details: map[string]interface{}{"field": "name"},
	}

	data, err := json.Marshal(&original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"code":"INVALID_INPUT"`)
	assert.Contains(t, string(data), `"message":"bad payload"`)

	var restored ErrorResponse
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, original.Status, restored.Status)
	assert.Equal(t, original.Code, restored.Code)
	assert.Equal(t, original.Message, restored.Message)
}

// Verifies: SYS-REQ-015
func TestErrorResponse_JSON_OmitsEmptyCode(t *testing.T) {
	e := ErrorResponse{Status: 500, Message: "server error"}
	data, err := json.Marshal(&e)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"code"`)
	assert.NotContains(t, string(data), `"details"`)
}
