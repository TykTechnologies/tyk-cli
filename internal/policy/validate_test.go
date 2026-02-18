package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

func validPolicyFile() types.PolicyFile {
	return types.PolicyFile{
		APIVersion: "tyk.tyktech/v1",
		Kind:       "Policy",
		Metadata: types.PolicyMetadata{
			ID:   "gold",
			Name: "Gold Plan",
		},
		Spec: types.PolicySpec{
			RateLimit: &types.RateLimit{Requests: 1000, Per: "60"},
			Quota:     &types.Quota{Limit: 100000, Period: "30d"},
			KeyTTL:    "0",
			Access: []types.AccessEntry{
				{Name: "users-api", Versions: []string{"v1"}},
			},
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
			"missing metadata.id",
			func(pf *types.PolicyFile) { pf.Metadata.ID = "" },
			"metadata.id", "schema",
		},
		{
			"missing metadata.name",
			func(pf *types.PolicyFile) { pf.Metadata.Name = "" },
			"metadata.name", "schema",
		},
		{
			"missing apiVersion",
			func(pf *types.PolicyFile) { pf.APIVersion = "" },
			"apiVersion", "schema",
		},
		{
			"wrong apiVersion",
			func(pf *types.PolicyFile) { pf.APIVersion = "wrong/v1" },
			"apiVersion", "schema",
		},
		{
			"missing kind",
			func(pf *types.PolicyFile) { pf.Kind = "" },
			"kind", "schema",
		},
		{
			"wrong kind",
			func(pf *types.PolicyFile) { pf.Kind = "Deployment" },
			"kind", "schema",
		},
		{
			"empty access list",
			func(pf *types.PolicyFile) { pf.Spec.Access = nil },
			"spec.access", "schema",
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
			func(pf *types.PolicyFile) { pf.Spec.RateLimit.Per = "abc" },
			"spec.rateLimit.per",
		},
		{
			"invalid quota.period",
			func(pf *types.PolicyFile) { pf.Spec.Quota.Period = "1.5h" },
			"spec.quota.period",
		},
		{
			"invalid keyTTL",
			func(pf *types.PolicyFile) { pf.Spec.KeyTTL = "-1" },
			"spec.keyTTL",
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
			pf.Spec.Access = []types.AccessEntry{tt.entry}
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
		// Missing apiVersion, kind, metadata.id, metadata.name
		Spec: types.PolicySpec{
			RateLimit: &types.RateLimit{Requests: 1000, Per: "abc"},
			Access: []types.AccessEntry{
				{Name: "foo", ID: "bar"}, // multiple selectors
			},
		},
	}

	errs := ValidatePolicy(pf)
	// Should have at least: apiVersion, kind, metadata.id, metadata.name, duration, selector = 6
	assert.GreaterOrEqual(t, len(errs), 6,
		"expected at least 6 errors for multiply-broken policy, got %d: %v", len(errs), errs)
}
