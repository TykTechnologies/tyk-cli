package policy

import (
	"fmt"

	"github.com/tyktech/tyk-cli/pkg/types"
)

const (
	requiredAPIVersion = "tyk.tyktech/v1"
	requiredKind       = "Policy"
)

// ValidatePolicy validates a PolicyFile and collects all errors before returning.
// It checks schema (required fields, types), duration formats, and selector constraints.
func ValidatePolicy(pf types.PolicyFile) types.ValidationErrors {
	var errs types.ValidationErrors

	// Schema: required fields
	if pf.APIVersion == "" {
		errs = append(errs, types.ValidationError{
			Field: "apiVersion", Message: "required field missing", Kind: "schema",
		})
	} else if pf.APIVersion != requiredAPIVersion {
		errs = append(errs, types.ValidationError{
			Field: "apiVersion", Message: fmt.Sprintf("must be %q", requiredAPIVersion), Kind: "schema",
		})
	}

	if pf.Kind == "" {
		errs = append(errs, types.ValidationError{
			Field: "kind", Message: "required field missing", Kind: "schema",
		})
	} else if pf.Kind != requiredKind {
		errs = append(errs, types.ValidationError{
			Field: "kind", Message: fmt.Sprintf("must be %q", requiredKind), Kind: "schema",
		})
	}

	if pf.Metadata.ID == "" {
		errs = append(errs, types.ValidationError{
			Field: "metadata.id", Message: "required field missing", Kind: "schema",
		})
	}

	if pf.Metadata.Name == "" {
		errs = append(errs, types.ValidationError{
			Field: "metadata.name", Message: "required field missing", Kind: "schema",
		})
	}

	if len(pf.Spec.Access) == 0 {
		errs = append(errs, types.ValidationError{
			Field: "spec.access", Message: "at least one access entry required", Kind: "schema",
		})
	}

	// Duration validation
	if pf.Spec.RateLimit != nil {
		if pf.Spec.RateLimit.Per != "" {
			if _, err := ParseDuration(string(pf.Spec.RateLimit.Per)); err != nil {
				errs = append(errs, types.ValidationError{
					Field: "spec.rateLimit.per", Message: err.Error(), Kind: "duration",
				})
			}
		}
	}

	if pf.Spec.Quota != nil {
		if pf.Spec.Quota.Period != "" {
			if _, err := ParseDuration(string(pf.Spec.Quota.Period)); err != nil {
				errs = append(errs, types.ValidationError{
					Field: "spec.quota.period", Message: err.Error(), Kind: "duration",
				})
			}
		}
	}

	if pf.Spec.KeyTTL != "" {
		if _, err := ParseDuration(string(pf.Spec.KeyTTL)); err != nil {
			errs = append(errs, types.ValidationError{
				Field: "spec.keyTTL", Message: err.Error(), Kind: "duration",
			})
		}
	}

	// Selector constraints per access entry
	for i, entry := range pf.Spec.Access {
		count := selectorCount(entry)
		if count != 1 {
			errs = append(errs, types.ValidationError{
				Field:   fmt.Sprintf("spec.access[%d]", i),
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
