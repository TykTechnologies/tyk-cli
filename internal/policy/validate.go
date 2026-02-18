package policy

import (
	"fmt"
	"regexp"

	"github.com/tyktech/tyk-cli/pkg/types"
)

var friendlyIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// ValidatePolicy validates a PolicyFile and collects all errors before returning.
// It checks schema (required fields, types), duration formats, and selector constraints.
func ValidatePolicy(pf types.PolicyFile) types.ValidationErrors {
	var errs types.ValidationErrors

	// Schema: required fields
	if pf.ID == "" {
		errs = append(errs, types.ValidationError{
			Field: "id", Message: "required field missing", Kind: "schema",
		})
	} else if verr := validateFriendlyID(pf.ID); verr != nil {
		errs = append(errs, *verr)
	}

	if pf.Name == "" {
		errs = append(errs, types.ValidationError{
			Field: "name", Message: "required field missing", Kind: "schema",
		})
	}

	if len(pf.Access) == 0 {
		errs = append(errs, types.ValidationError{
			Field: "access", Message: "at least one access entry required", Kind: "schema",
		})
	}

	// Duration validation
	if pf.RateLimit != nil {
		if pf.RateLimit.Per != "" {
			if _, err := ParseDuration(string(pf.RateLimit.Per)); err != nil {
				errs = append(errs, types.ValidationError{
					Field: "rateLimit.per", Message: err.Error(), Kind: "duration",
				})
			}
		}
	}

	if pf.Quota != nil {
		if pf.Quota.Period != "" {
			if _, err := ParseDuration(string(pf.Quota.Period)); err != nil {
				errs = append(errs, types.ValidationError{
					Field: "quota.period", Message: err.Error(), Kind: "duration",
				})
			}
		}
	}

	if pf.KeyTTL != "" {
		if _, err := ParseDuration(string(pf.KeyTTL)); err != nil {
			errs = append(errs, types.ValidationError{
				Field: "keyTTL", Message: err.Error(), Kind: "duration",
			})
		}
	}

	// Selector constraints per access entry
	for i, entry := range pf.Access {
		count := selectorCount(entry)
		if count != 1 {
			errs = append(errs, types.ValidationError{
				Field:   fmt.Sprintf("access[%d]", i),
				Message: "exactly one of id, name, listenPath, or tags must be set",
				Kind:    "selector",
			})
		}
	}

	return errs
}

// validateFriendlyID validates the format of a friendly policy ID.
func validateFriendlyID(id string) *types.ValidationError {
	if len(id) > 64 {
		return &types.ValidationError{Field: "id", Message: "must be 64 characters or fewer", Kind: "schema"}
	}
	if !friendlyIDPattern.MatchString(id) {
		return &types.ValidationError{
			Field:   "id",
			Message: "must contain only lowercase letters, numbers, dots, hyphens, underscores, and start with a letter or number",
			Kind:    "schema",
		}
	}
	if isObjectIDFormat(id) {
		return &types.ValidationError{
			Field:   "id",
			Message: "looks like a MongoDB ObjectID — use a human-readable name instead (e.g., 'gold', 'free-tier')",
			Kind:    "schema",
		}
	}
	return nil
}

// isObjectIDFormat returns true if s looks like a 24-character hex MongoDB ObjectID.
func isObjectIDFormat(s string) bool {
	if len(s) != 24 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// selectorCount returns how many selector fields are set on an AccessEntry.
func selectorCount(e types.AccessEntry) int {
	count := 0
	if e.ID != "" {
		count++
	}
	if e.Name != "" {
		count++
	}
	if e.ListenPath != "" {
		count++
	}
	if len(e.Tags) > 0 {
		count++
	}
	return count
}
