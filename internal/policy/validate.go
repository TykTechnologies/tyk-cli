package policy

import (
	"fmt"

	"github.com/tyktech/tyk-cli/pkg/types"
)

// ValidatePolicy validates a PolicyFile and collects all errors before returning.
// It checks schema (required fields, types), duration formats, and selector constraints.
func ValidatePolicy(pf types.PolicyFile) types.ValidationErrors {
	var errs types.ValidationErrors

	// Schema: required fields
	if pf.ID == "" {
		errs = append(errs, types.ValidationError{
			Field: "id", Message: "required field missing", Kind: "schema",
		})
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
