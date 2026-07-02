package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Verifies: SYS-REQ-029
func validPolicyFile() types.PolicyFile {
	return types.PolicyFile{
		ID:   "gold",
		Name: "Gold Plan",
		RateLimit: &types.RateLimit{Requests: 1000, Per: "60"},
		Quota:     &types.Quota{Limit: 100000, Period: "30d"},
		KeyTTL:    "0",
		Access: []types.AccessEntry{
			{Name: "users-api", Versions: []string{"v1"}},
		},
	}
}

// Verifies: SYS-REQ-029
func TestValidatePolicy_Valid(t *testing.T) {
	errs := ValidatePolicy(validPolicyFile())
	assert.Empty(t, errs)
}

// Verifies: SYS-REQ-029
func TestValidatePolicy_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name          string
		modify        func(*types.PolicyFile)
		expectedField string
		expectedKind  string
	}{
		{
			"missing id",
			func(pf *types.PolicyFile) { pf.ID = "" },
			"id", "schema",
		},
		{
			"missing name",
			func(pf *types.PolicyFile) { pf.Name = "" },
			"name", "schema",
		},
		{
			"empty access list",
			func(pf *types.PolicyFile) { pf.Access = nil },
			"access", "schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := validPolicyFile()
			tt.modify(&pf)
			errs := ValidatePolicy(pf)
			require.NotEmpty(t, errs, "expected validation error for %s", tt.name)

			found := false
			for _, e := range errs {
				if e.Field == tt.expectedField && e.Kind == tt.expectedKind {
					found = true
					break
				}
			}
			assert.True(t, found, "expected error for field %s with kind %s, got: %v",
				tt.expectedField, tt.expectedKind, errs)
		})
	}
}

// Verifies: SYS-REQ-029
func TestValidatePolicy_InvalidDurations(t *testing.T) {
	tests := []struct {
		name          string
		modify        func(*types.PolicyFile)
		expectedField string
	}{
		{
			"invalid rateLimit.per",
			func(pf *types.PolicyFile) { pf.RateLimit.Per = "abc" },
			"rateLimit.per",
		},
		{
			"invalid quota.period",
			func(pf *types.PolicyFile) { pf.Quota.Period = "1.5h" },
			"quota.period",
		},
		{
			"invalid keyTTL",
			func(pf *types.PolicyFile) { pf.KeyTTL = "-1" },
			"keyTTL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := validPolicyFile()
			tt.modify(&pf)
			errs := ValidatePolicy(pf)
			require.NotEmpty(t, errs)

			found := false
			for _, e := range errs {
				if e.Field == tt.expectedField && e.Kind == "duration" {
					found = true
					break
				}
			}
			assert.True(t, found, "expected duration error for field %s, got: %v",
				tt.expectedField, errs)
		})
	}
}

// Verifies: SYS-REQ-029
func TestValidatePolicy_SelectorConstraints(t *testing.T) {
	tests := []struct {
		name   string
		entry  types.AccessEntry
		errMsg string
	}{
		{
			"zero selectors",
			types.AccessEntry{Versions: []string{"v1"}},
			"exactly one of",
		},
		{
			"multiple selectors: name and id",
			types.AccessEntry{Name: "foo", ID: "bar"},
			"exactly one of",
		},
		{
			"multiple selectors: name and listenPath",
			types.AccessEntry{Name: "foo", ListenPath: "/bar/"},
			"exactly one of",
		},
		{
			"empty tags slice",
			types.AccessEntry{Tags: []string{}},
			"exactly one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := validPolicyFile()
			pf.Access = []types.AccessEntry{tt.entry}
			errs := ValidatePolicy(pf)
			require.NotEmpty(t, errs)

			found := false
			for _, e := range errs {
				if e.Kind == "selector" {
					found = true
					break
				}
			}
			assert.True(t, found, "expected selector error, got: %v", errs)
		})
	}
}

// Verifies: SYS-REQ-029
func TestValidatePolicy_CollectsAllErrors(t *testing.T) {
	pf := types.PolicyFile{
		// Missing id, name
		RateLimit: &types.RateLimit{Requests: 1000, Per: "abc"},
		Access: []types.AccessEntry{
			{Name: "foo", ID: "bar"}, // multiple selectors
		},
	}

	errs := ValidatePolicy(pf)
	// Should have at least: id, name, duration, selector = 4
	assert.GreaterOrEqual(t, len(errs), 4,
		"expected at least 4 errors for multiply-broken policy, got %d: %v", len(errs), errs)
}

// Verifies: SYS-REQ-029
func TestValidatePolicy_FriendlyID_Valid(t *testing.T) {
	validIDs := []string{"gold", "free-tier", "rate-limit-basic", "v2.0", "a", "abc_def"}

	for _, id := range validIDs {
		t.Run(id, func(t *testing.T) {
			pf := validPolicyFile()
			pf.ID = id
			errs := ValidatePolicy(pf)
			for _, e := range errs {
				assert.NotEqual(t, "id", e.Field, "expected no id error for valid ID %q, got: %v", id, e)
			}
		})
	}
}

// Verifies: SYS-REQ-029
func TestValidatePolicy_FriendlyID_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantMsg string
	}{
		{"uppercase", "Gold", "lowercase"},
		{"starts with hyphen", "-bad", "start with"},
		{"special chars", "gold!", "lowercase"},
		{"too long", "aaaaaaaaaabbbbbbbbbbccccccccccddddddddddeeeeeeeeeeffffffffffggggg", "64 characters"},
		{"ObjectID-shaped", "507f1f77bcf86cd799439011", "MongoDB ObjectID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := validPolicyFile()
			pf.ID = tt.id
			errs := ValidatePolicy(pf)
			require.NotEmpty(t, errs, "expected validation error for ID %q", tt.id)

			found := false
			for _, e := range errs {
				if e.Field == "id" && e.Kind == "schema" {
					found = true
					assert.Contains(t, e.Message, tt.wantMsg,
						"error message for ID %q should contain %q", tt.id, tt.wantMsg)
					break
				}
			}
			assert.True(t, found, "expected schema error for field 'id', got: %v", errs)
		})
	}
}

// ===========================================================================
// MC/DC coverage of ValidatePolicy duration branches and isObjectIDFormat
// character-class condition independence.
// ===========================================================================

// Verifies: SYS-REQ-029
func TestValidatePolicy_DurationBranches_MCDC(t *testing.T) {
	tests := []struct {
		name           string
		modify         func(*types.PolicyFile)
		expectField    string
		expectKindDur  bool
	}{
		{
			name:          "rateLimit.per present and invalid",
			modify:        func(p *types.PolicyFile) { p.RateLimit = &types.RateLimit{Requests: 100, Per: "garbage"} },
			expectField:   "rateLimit.per",
			expectKindDur: true,
		},
		{
			name:          "quota.period present and invalid",
			modify:        func(p *types.PolicyFile) { p.Quota = &types.Quota{Limit: 1000, Period: "garbage"} },
			expectField:   "quota.period",
			expectKindDur: true,
		},
		{
			name:          "keyTTL present and invalid",
			modify:        func(p *types.PolicyFile) { p.KeyTTL = "garbage" },
			expectField:   "keyTTL",
			expectKindDur: true,
		},
		{
			name: "rateLimit nil — branch skipped",
			modify: func(p *types.PolicyFile) {
				p.RateLimit = nil
			},
			expectField:   "",
			expectKindDur: false,
		},
		{
			name: "rateLimit non-nil but Per empty — inner branch skipped",
			modify: func(p *types.PolicyFile) {
				p.RateLimit = &types.RateLimit{Requests: 100, Per: ""}
			},
			expectField:   "",
			expectKindDur: false,
		},
		{
			name: "quota nil — branch skipped",
			modify: func(p *types.PolicyFile) {
				p.Quota = nil
			},
			expectField:   "",
			expectKindDur: false,
		},
		{
			name: "quota non-nil but Period empty — inner branch skipped",
			modify: func(p *types.PolicyFile) {
				p.Quota = &types.Quota{Limit: 1000, Period: ""}
			},
			expectField:   "",
			expectKindDur: false,
		},
		{
			name: "keyTTL empty — branch skipped",
			modify: func(p *types.PolicyFile) {
				p.KeyTTL = ""
			},
			expectField:   "",
			expectKindDur: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := validPolicyFile()
			tt.modify(&pf)
			errs := ValidatePolicy(pf)

			if tt.expectKindDur {
				found := false
				for _, e := range errs {
					if e.Field == tt.expectField && e.Kind == "duration" {
						found = true
						break
					}
				}
				assert.True(t, found, "expected duration error on %s, got: %v", tt.expectField, errs)
			} else {
				for _, e := range errs {
					assert.NotEqual(t, "duration", e.Kind, "no duration error expected, got: %v", e)
				}
			}
		})
	}
}

// Verifies: SYS-REQ-029
// MC/DC: selectorCount's len(e.Tags) > 0 = T branch (validate.go:128).
// A policy with Tags as the sole selector must validate without selector errors,
// proving the count++ branch executes when Tags is non-empty.
func TestValidatePolicy_TagsAsSoleSelector(t *testing.T) {
	pf := validPolicyFile()
	pf.Access = []types.AccessEntry{
		{Tags: []string{"public"}, Versions: []string{"v1"}},
	}
	errs := ValidatePolicy(pf)
	for _, e := range errs {
		assert.NotEqual(t, "selector", e.Kind,
			"expected no selector error when Tags is the sole selector, got: %v", e)
	}
}

// Verifies: SYS-REQ-029
// isObjectIDFormat checks: len == 24 AND every char in [0-9a-f].
// MC/DC requires independent exercise of each character-class condition.
func TestIsObjectIDFormat_MCDC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", false},
		{"short hex", "abc123", false},
		{"24 chars all hex", "507f1f77bcf86cd799439011", true},
		{"24 chars with uppercase hex", "507F1F77BCF86CD799439011", false},
		{"24 chars with G (out of hex)", "g07f1f77bcf86cd799439011", false},
		{"24 chars with digit only", "123456789012345678901234", true},
		{"24 chars with letters a-f only", "abcdefabcdefabcdefabcdef", true},
		{"24 chars one non-hex letter", "abcdefabcdefabcdefabcde!", false},
		{"23 chars", "abcdefabcdefabcdefabcde", false},
		{"25 chars", "abcdefabcdefabcdefabcdeff", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isObjectIDFormat(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
