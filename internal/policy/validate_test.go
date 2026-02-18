package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

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

func TestValidatePolicy_Valid(t *testing.T) {
	errs := ValidatePolicy(validPolicyFile())
	assert.Empty(t, errs)
}

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
