package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyktech/tyk-cli/internal/cli"
)

// Verifies: SYS-REQ-013
func TestClassifyExitError_NilErrorReturnsZero(t *testing.T) {
	var buf bytes.Buffer
	code := classifyExitError(nil, &buf)
	assert.Equal(t, 0, code, "nil error must map to exit code 0")
	assert.Empty(t, buf.String(), "nothing should be written to stderr when err is nil")
}

// Verifies: SYS-REQ-013
func TestClassifyExitError_ExitErrorPreservesCodeAndMessage(t *testing.T) {
	tests := []struct {
		name    string
		err     *cli.ExitError
		wantCode int
		wantMsg  string
	}{
		{"bad args (2)", &cli.ExitError{Code: 2, Message: "bad arguments"}, 2, "bad arguments"},
		{"not found (3)", &cli.ExitError{Code: 3, Message: "API 'gold' not found"}, 3, "API 'gold' not found"},
		{"conflict (4)", &cli.ExitError{Code: 4, Message: "conflict on listen path"}, 4, "conflict on listen path"},
		{"auth failed (5)", &cli.ExitError{Code: 5, Message: "authentication failed"}, 5, "authentication failed"},
		{"forbidden (6)", &cli.ExitError{Code: 6, Message: "forbidden"}, 6, "forbidden"},
		{"rate limited (7)", &cli.ExitError{Code: 7, Message: "rate limited"}, 7, "rate limited"},
		{"server error (8)", &cli.ExitError{Code: 8, Message: "server error"}, 8, "server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			code := classifyExitError(tt.err, &buf)
			assert.Equal(t, tt.wantCode, code, "exit code must match the ExitError code")
			assert.Contains(t, buf.String(), tt.wantMsg, "stderr must contain the ExitError message")
			assert.Contains(t, buf.String(), "Error:", "stderr line must start with the 'Error:' prefix")
		})
	}
}

// Verifies: SYS-REQ-013
func TestClassifyExitError_GenericErrorFallsBackToExitOne(t *testing.T) {
	var buf bytes.Buffer
	err := errors.New("some unexpected failure")
	code := classifyExitError(err, &buf)
	assert.Equal(t, 1, code, "generic non-ExitError must fall through to exit 1")
	assert.Contains(t, buf.String(), "some unexpected failure",
		"stderr must include the original error message")
}

// Verifies: SYS-REQ-013
// classifyExitError must follow `errors.As` semantics: an ExitError wrapped
// inside another error chain must still be detected.
func TestClassifyExitError_WrappedExitErrorIsDetected(t *testing.T) {
	wrapped := fmt.Errorf("while processing: %w", &cli.ExitError{Code: 4, Message: "inner conflict"})
	var buf bytes.Buffer
	code := classifyExitError(wrapped, &buf)
	assert.Equal(t, 4, code, "wrapped ExitError must still yield its code")
	assert.Contains(t, buf.String(), "inner conflict",
		"stderr must reference the inner ExitError message")
}
