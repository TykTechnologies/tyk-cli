package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionNotFoundError(t *testing.T) {
	err := versionNotFoundError("abc123", "Petstore", "v99", []string{"v1", "v2", "v3"})

	assert.Equal(t, 3, err.Code)
	assert.Contains(t, err.Message, `version "v99" not found for API abc123 (Petstore)`)
	assert.Contains(t, err.Message, "Available versions: v1, v2, v3")
}

func TestVersionConflictError(t *testing.T) {
	err := versionConflictError("abc123", "v2")

	assert.Equal(t, 4, err.Code)
	assert.Contains(t, err.Message, `version "v2" already exists for API abc123`)
	assert.Contains(t, err.Message, "tyk api apply -f <file>")
}

func TestAPINotFoundForVersionError(t *testing.T) {
	err := apiNotFoundForVersionError("abc123")

	assert.Equal(t, 3, err.Code)
	assert.Contains(t, err.Message, "API not found: abc123")
	assert.Contains(t, err.Message, "tyk api list")
}
