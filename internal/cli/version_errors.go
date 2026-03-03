package cli

import (
	"fmt"
	"strings"
)

func versionNotFoundError(apiID, apiName, requestedVersion string, availableVersions []string) *ExitError {
	return &ExitError{
		Code:    3,
		Message: fmt.Sprintf(`version "%s" not found for API %s (%s). Available versions: %s.`, requestedVersion, apiID, apiName, strings.Join(availableVersions, ", ")),
	}
}

func versionConflictError(apiID, versionName string) *ExitError {
	return &ExitError{
		Code:    4,
		Message: fmt.Sprintf(`version "%s" already exists for API %s. Use 'tyk api apply -f <file>' to update.`, versionName, apiID),
	}
}

func apiNotFoundForVersionError(apiID string) *ExitError {
	return &ExitError{
		Code:    3,
		Message: fmt.Sprintf("API not found: %s. Use 'tyk api list' to see available APIs.", apiID),
	}
}
