package cli

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tyktech/tyk-cli/pkg/types"
)

// ExitError represents an error with a specific exit code
type ExitError struct {
	Code    int
	Message string
}

// reqproof:req REQ-API-020
func (e *ExitError) Error() string {
	return e.Message
}

// httpStatusFromError extracts the HTTP status code from a Dashboard error
// response when one is available. Returns 0 for non-HTTP errors.
//
// reqproof:req REQ-API-024
// reqproof:req REQ-POL-024
func httpStatusFromError(err error) int {
	if err == nil {
		return 0
	}
	var er *types.ErrorResponse
	if errors.As(err, &er) && er != nil {
		return er.Status
	}
	return 0
}

// isConflictError reports whether err represents an HTTP 409 conflict from the
// Dashboard. It checks the typed *types.ErrorResponse first (the canonical
// signal) and falls back to a substring match for wrapped/string-only errors.
//
// reqproof:req REQ-API-023
// reqproof:req REQ-POL-023
func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	if httpStatusFromError(err) == http.StatusConflict {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "409") || strings.Contains(msg, "conflict")
}

// isAuthError reports whether err is an HTTP 401 from the Dashboard.
//
// reqproof:req REQ-API-024
// reqproof:req REQ-POL-024
func isAuthError(err error) bool {
	return httpStatusFromError(err) == http.StatusUnauthorized
}

// isForbiddenError reports whether err is an HTTP 403 from the Dashboard.
//
// reqproof:req REQ-API-024
// reqproof:req REQ-POL-024
func isForbiddenError(err error) bool {
	return httpStatusFromError(err) == http.StatusForbidden
}

// isRateLimitError reports whether err is an HTTP 429 from the Dashboard.
//
// reqproof:req REQ-API-025
// reqproof:req REQ-POL-025
func isRateLimitError(err error) bool {
	return httpStatusFromError(err) == http.StatusTooManyRequests
}

// isServerError reports whether err is a 5xx response from the Dashboard.
//
// reqproof:req REQ-API-026
// reqproof:req REQ-POL-026
func isServerError(err error) bool {
	s := httpStatusFromError(err)
	return s >= 500 && s <= 599
}

// classifyDashboardError maps an HTTP-error result to its specific ExitError
// when one of the canonical Dashboard failure classes applies. Returns nil
// when the error does not match any classified status (caller falls through
// to its existing wrap/ExitError{Code: 1} path).
//
// reqproof:req REQ-API-024
// reqproof:req REQ-API-025
// reqproof:req REQ-API-026
// reqproof:req REQ-API-027
// reqproof:req REQ-POL-022
// reqproof:req REQ-POL-024
// reqproof:req REQ-POL-025
// reqproof:req REQ-POL-026
// reqproof:req REQ-POL-027
func classifyDashboardError(err error, op string) *ExitError {
	if err == nil {
		return nil
	}
	switch {
	case isAuthError(err):
		return &ExitError{Code: int(types.ExitAuthFailed), Message: fmt.Sprintf("%s: authentication failed (HTTP 401); rotate the auth token", op)}
	case isForbiddenError(err):
		return &ExitError{Code: int(types.ExitForbidden), Message: fmt.Sprintf("%s: forbidden (HTTP 403); the token lacks the required permission", op)}
	case isRateLimitError(err):
		return &ExitError{Code: int(types.ExitRateLimited), Message: fmt.Sprintf("%s: rate-limited by Dashboard (HTTP 429); back off and retry", op)}
	case isServerError(err):
		return &ExitError{Code: int(types.ExitServerError), Message: fmt.Sprintf("%s: Dashboard returned a server error (%v); retry later", op, err)}
	}
	return nil
}