package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Verifies: SYS-REQ-030
func TestPolicyFile_YAMLRoundTrip(t *testing.T) {
	original := PolicyFile{
		ID:   "gold",
		Name: "Gold Plan",
		Tags: []string{"gold", "paid"},
		RateLimit: &RateLimit{Requests: 1000, Per: Duration("1m")},
		Quota:     &Quota{Limit: 100000, Period: Duration("30d")},
		KeyTTL:    Duration("0"),
		Access: []AccessEntry{
			{Name: "users-api", Versions: []string{"v1"}},
			{ListenPath: "/orders/", Versions: []string{"v1", "v2"}},
			{Tags: []string{"public", "v1"}, Versions: []string{"v1"}},
			{ID: "foobar123"},
		},
	}

	yamlBytes, err := yaml.Marshal(&original)
	require.NoError(t, err)

	var restored PolicyFile
	err = yaml.Unmarshal(yamlBytes, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.Tags, restored.Tags)
	assert.Equal(t, original.RateLimit.Requests, restored.RateLimit.Requests)
	assert.Equal(t, original.RateLimit.Per, restored.RateLimit.Per)
	assert.Equal(t, original.Quota.Limit, restored.Quota.Limit)
	assert.Equal(t, original.Quota.Period, restored.Quota.Period)
	assert.Equal(t, original.KeyTTL, restored.KeyTTL)
	require.Len(t, restored.Access, 4)
	assert.Equal(t, "users-api", restored.Access[0].Name)
	assert.Equal(t, []string{"v1"}, restored.Access[0].Versions)
	assert.Equal(t, "/orders/", restored.Access[1].ListenPath)
	assert.Equal(t, []string{"v1", "v2"}, restored.Access[1].Versions)
	assert.Equal(t, []string{"public", "v1"}, restored.Access[2].Tags)
	assert.Equal(t, "foobar123", restored.Access[3].ID)
	assert.Empty(t, restored.Access[3].Versions, "omitted versions should remain nil/empty")
}

// Verifies: SYS-REQ-030
func TestDashboardPolicy_JSONRoundTrip(t *testing.T) {
	wireJSON := `{
		"_id": "gold",
		"id": "",
		"name": "Gold Plan",
		"org_id": "5e9d9544a1dcd60001d0ed20",
		"rate": 1000,
		"per": 60,
		"quota_max": 100000,
		"quota_renewal_rate": 2592000,
		"key_expires_in": 0,
		"tags": ["gold", "paid"],
		"access_rights": {
			"a1b2c3d4e5f6": {
				"api_id": "a1b2c3d4e5f6",
				"api_name": "users-api",
				"versions": ["v1"],
				"allowed_urls": [],
				"limit": null
			},
			"g7h8i9j0k1l2": {
				"api_id": "g7h8i9j0k1l2",
				"api_name": "orders-api",
				"versions": ["v1", "v2"],
				"allowed_urls": [],
				"limit": null
			}
		},
		"active": true,
		"is_inactive": false
	}`

	var policy DashboardPolicy
	err := json.Unmarshal([]byte(wireJSON), &policy)
	require.NoError(t, err)

	assert.Equal(t, "gold", policy.MID)
	assert.Equal(t, "", policy.ID)
	assert.Equal(t, "Gold Plan", policy.Name)
	assert.Equal(t, "5e9d9544a1dcd60001d0ed20", policy.OrgID)
	assert.Equal(t, int64(1000), policy.Rate)
	assert.Equal(t, int64(60), policy.Per)
	assert.Equal(t, int64(100000), policy.QuotaMax)
	assert.Equal(t, int64(2592000), policy.QuotaRenewalRate)
	assert.Equal(t, int64(0), policy.KeyExpiresIn)
	assert.Equal(t, []string{"gold", "paid"}, policy.Tags)
	assert.True(t, policy.Active)
	assert.False(t, policy.IsInactive)

	require.Len(t, policy.AccessRights, 2)
	usersRight := policy.AccessRights["a1b2c3d4e5f6"]
	assert.Equal(t, "a1b2c3d4e5f6", usersRight.APIID)
	assert.Equal(t, "users-api", usersRight.APIName)
	assert.Equal(t, []string{"v1"}, usersRight.Versions)

	ordersRight := policy.AccessRights["g7h8i9j0k1l2"]
	assert.Equal(t, "g7h8i9j0k1l2", ordersRight.APIID)
	assert.Equal(t, "orders-api", ordersRight.APIName)
	assert.Equal(t, []string{"v1", "v2"}, ordersRight.Versions)

	// Re-marshal and unmarshal to verify round-trip
	remarshaled, err := json.Marshal(&policy)
	require.NoError(t, err)

	var roundTripped DashboardPolicy
	err = json.Unmarshal(remarshaled, &roundTripped)
	require.NoError(t, err)

	assert.Equal(t, policy.MID, roundTripped.MID)
	assert.Equal(t, policy.Name, roundTripped.Name)
	assert.Equal(t, policy.Rate, roundTripped.Rate)
	assert.Equal(t, policy.Per, roundTripped.Per)
	assert.Equal(t, policy.QuotaMax, roundTripped.QuotaMax)
	assert.Equal(t, policy.QuotaRenewalRate, roundTripped.QuotaRenewalRate)
	assert.Equal(t, policy.AccessRights, roundTripped.AccessRights)
}

// Verifies: SYS-REQ-031
func TestDuration_UnmarshalYAML(t *testing.T) {
	t.Run("string durations", func(t *testing.T) {
		tests := []struct {
			input    string
			expected Duration
		}{
			{`per: "30d"`, Duration("30d")},
			{`per: "1m"`, Duration("1m")},
			{`per: "24h"`, Duration("24h")},
			{`per: "60s"`, Duration("60s")},
			{`per: "0"`, Duration("0")},
		}
		for _, tt := range tests {
			var dest struct {
				Per Duration `yaml:"per"`
			}
			err := yaml.Unmarshal([]byte(tt.input), &dest)
			require.NoError(t, err, "input: %s", tt.input)
			assert.Equal(t, tt.expected, dest.Per, "input: %s", tt.input)
		}
	})

	t.Run("integer durations", func(t *testing.T) {
		tests := []struct {
			input    string
			expected Duration
		}{
			{`per: 60`, Duration("60")},
			{`per: 0`, Duration("0")},
			{`per: 2592000`, Duration("2592000")},
		}
		for _, tt := range tests {
			var dest struct {
				Per Duration `yaml:"per"`
			}
			err := yaml.Unmarshal([]byte(tt.input), &dest)
			require.NoError(t, err, "input: %s", tt.input)
			assert.Equal(t, tt.expected, dest.Per, "input: %s", tt.input)
		}
	})
}

// Verifies: SYS-REQ-032
func TestAccessEntry_SelectorFields(t *testing.T) {
	// Verify that each selector field is independently settable and
	// survives YAML round-trip in isolation.
	tests := []struct {
		name  string
		yaml  string
		check func(t *testing.T, e AccessEntry)
	}{
		{
			name: "id selector only",
			yaml: "id: abc123\nversions: [v1]",
			check: func(t *testing.T, e AccessEntry) {
				assert.Equal(t, "abc123", e.ID)
				assert.Empty(t, e.Name)
				assert.Empty(t, e.ListenPath)
				assert.Empty(t, e.Tags)
				assert.Equal(t, []string{"v1"}, e.Versions)
			},
		},
		{
			name: "name selector only",
			yaml: "name: users-api",
			check: func(t *testing.T, e AccessEntry) {
				assert.Empty(t, e.ID)
				assert.Equal(t, "users-api", e.Name)
				assert.Empty(t, e.ListenPath)
				assert.Empty(t, e.Tags)
			},
		},
		{
			name: "listenPath selector only",
			yaml: "listenPath: /orders/",
			check: func(t *testing.T, e AccessEntry) {
				assert.Empty(t, e.ID)
				assert.Empty(t, e.Name)
				assert.Equal(t, "/orders/", e.ListenPath)
				assert.Empty(t, e.Tags)
			},
		},
		{
			name: "tags selector only",
			yaml: "tags: [public, v1]",
			check: func(t *testing.T, e AccessEntry) {
				assert.Empty(t, e.ID)
				assert.Empty(t, e.Name)
				assert.Empty(t, e.ListenPath)
				assert.Equal(t, []string{"public", "v1"}, e.Tags)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entry AccessEntry
			err := yaml.Unmarshal([]byte(tt.yaml), &entry)
			require.NoError(t, err)
			tt.check(t, entry)
		})
	}
}

// Verifies: SYS-REQ-030
func TestAccessRight_MarshalJSON_NilHandling(t *testing.T) {
	t.Run("nil AllowedURLs serializes as empty array", func(t *testing.T) {
		ar := AccessRight{APIID: "a1", APIName: "test", AllowedURLs: nil, Limit: nil}
		data, err := json.Marshal(&ar)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"allowed_urls":[]`)
		assert.Contains(t, string(data), `"limit":null`)
	})

	t.Run("non-nil AllowedURLs preserved", func(t *testing.T) {
		ar := AccessRight{
			APIID:       "a1",
			APIName:     "test",
			AllowedURLs: []AllowedURL{{URL: "/foo", Methods: []string{"GET"}}},
			Limit:       &RateQuotaLimit{Rate: 10, Per: 60},
		}
		data, err := json.Marshal(&ar)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"/foo"`)
		assert.NotContains(t, string(data), `"allowed_urls":[]`)
		assert.NotContains(t, string(data), `"limit":null`)
	})
}

// Verifies: SYS-REQ-029
func TestValidationError_Error(t *testing.T) {
	t.Run("formats field, message, and kind", func(t *testing.T) {
		e := &ValidationError{Field: "rateLimit.requests", Message: "must be positive", Kind: "schema"}
		assert.Equal(t, "rateLimit.requests: must be positive (schema)", e.Error())
	})

	t.Run("handles empty fields without panicking", func(t *testing.T) {
		e := &ValidationError{}
		assert.Equal(t, ":  ()", e.Error())
	})
}

// Verifies: SYS-REQ-029
func TestValidationErrors_Error(t *testing.T) {
	t.Run("empty slice returns sentinel message", func(t *testing.T) {
		// Covers the len(ve) == 0 == T branch.
		var ve ValidationErrors
		assert.Equal(t, "no validation errors", ve.Error())
	})

	t.Run("single error returns that error's message", func(t *testing.T) {
		// Covers len(ve) == 0 == F and len(ve) == 1 == T.
		ve := ValidationErrors{
			{Field: "access[0].id", Message: "must be set", Kind: "selector"},
		}
		assert.Equal(t, "access[0].id: must be set (selector)", ve.Error())
	})

	t.Run("multiple errors returns aggregated message", func(t *testing.T) {
		// Covers len(ve) == 0 == F and len(ve) == 1 == F.
		ve := ValidationErrors{
			{Field: "access[0].id", Message: "must be set", Kind: "selector"},
			{Field: "rateLimit.per", Message: "invalid duration", Kind: "duration"},
			{Field: "name", Message: "required", Kind: "schema"},
		}
		got := ve.Error()
		assert.Contains(t, got, "3 validation errors")
		assert.Contains(t, got, "access[0].id: must be set (selector)")
		assert.Contains(t, got, "and 2 more")
	})
}

// Verifies: SYS-REQ-024
func TestDashboardPolicyListResponse_JSONUnmarshal(t *testing.T) {
	listJSON := `{
		"Data": [
			{"_id": "gold", "name": "Gold Plan", "rate": 1000, "per": 60, "active": true},
			{"_id": "silver", "name": "Silver Plan", "rate": 500, "per": 60, "active": true}
		],
		"Pages": 1,
		"StatusCode": 200
	}`

	var resp DashboardPolicyListResponse
	err := json.Unmarshal([]byte(listJSON), &resp)
	require.NoError(t, err)

	assert.Equal(t, 1, resp.Pages)
	assert.Equal(t, 200, resp.StatusCode)
	require.Len(t, resp.Data, 2)
	assert.Equal(t, "gold", resp.Data[0].MID)
	assert.Equal(t, "Gold Plan", resp.Data[0].Name)
	assert.Equal(t, "silver", resp.Data[1].MID)
	assert.Equal(t, "Silver Plan", resp.Data[1].Name)
}
