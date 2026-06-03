package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// reqproof:req REQ-POL-007
func TestCLIToWire(t *testing.T) {
	resolved := []ResolvedAccess{
		{APIID: "a1b2c3d4e5f6", APIName: "users-api", Versions: []string{"v1"}},
		{APIID: "g7h8i9j0k1l2", APIName: "orders-api", Versions: []string{"v1", "v2"}},
	}

	pf := types.PolicyFile{
		ID:   "gold",
		Name: "Gold Plan",
		Tags: []string{"gold", "paid"},
		RateLimit: &types.RateLimit{Requests: 1000, Per: "60"},
		Quota:     &types.Quota{Limit: 100000, Period: "30d"},
		KeyTTL:    "0",
		Access: []types.AccessEntry{
			{Name: "users-api", Versions: []string{"v1"}},
			{ListenPath: "/orders/", Versions: []string{"v1", "v2"}},
		},
	}

	dp, err := CLIToWire(pf, resolved, "org-123")
	require.NoError(t, err)

	assert.Empty(t, dp.MID, "MID should be empty — caller sets it after resolution")
	assert.Equal(t, "gold", dp.ID, "wire id should match friendly id directly")
	assert.Equal(t, "Gold Plan", dp.Name)
	assert.Equal(t, "org-123", dp.OrgID)
	assert.Equal(t, []string{"gold", "paid"}, dp.Tags)
	assert.Equal(t, int64(1000), dp.Rate)
	assert.Equal(t, int64(60), dp.Per)
	assert.Equal(t, int64(100000), dp.QuotaMax)
	assert.Equal(t, int64(2592000), dp.QuotaRenewalRate)
	assert.Equal(t, int64(0), dp.KeyExpiresIn)
	assert.True(t, dp.Active)
	assert.False(t, dp.IsInactive)

	require.Len(t, dp.AccessRights, 2)
	usersAR := dp.AccessRights["a1b2c3d4e5f6"]
	require.NotNil(t, usersAR)
	assert.Equal(t, "a1b2c3d4e5f6", usersAR.APIID)
	assert.Equal(t, "users-api", usersAR.APIName)
	assert.Equal(t, []string{"v1"}, usersAR.Versions)

	ordersAR := dp.AccessRights["g7h8i9j0k1l2"]
	require.NotNil(t, ordersAR)
	assert.Equal(t, []string{"v1", "v2"}, ordersAR.Versions)
}

// reqproof:req REQ-POL-007
func TestWireToCLI(t *testing.T) {
	dp := types.DashboardPolicy{
		MID:              "507f1f77bcf86cd799439011",
		ID:               "gold",
		Name:             "Gold Plan",
		Tags:             []string{"gold", "paid"},
		Rate:             1000,
		Per:              60,
		QuotaMax:         100000,
		QuotaRenewalRate: 2592000,
		KeyExpiresIn:     0,
		Active:           true,
		AccessRights: map[string]*types.AccessRight{
			"a1b2c3d4e5f6": {
				APIID:   "a1b2c3d4e5f6",
				APIName: "users-api",
				Versions: []string{"v1"},
			},
			"g7h8i9j0k1l2": {
				APIID:   "g7h8i9j0k1l2",
				APIName: "orders-api",
				Versions: []string{"v1", "v2"},
			},
		},
	}

	apis := []ResolverAPI{
		{ID: "a1b2c3d4e5f6", Name: "users-api"},
		{ID: "g7h8i9j0k1l2", Name: "orders-api"},
	}

	pf := WireToCLI(dp, apis)

	assert.Equal(t, "gold", pf.ID, "should use wire id directly")
	assert.Equal(t, "Gold Plan", pf.Name)
	assert.Equal(t, []string{"gold", "paid"}, pf.Tags)
	assert.Equal(t, int64(1000), pf.RateLimit.Requests)
	assert.Equal(t, types.Duration("1m"), pf.RateLimit.Per)
	assert.Equal(t, int64(100000), pf.Quota.Limit)
	assert.Equal(t, types.Duration("30d"), pf.Quota.Period)
	assert.Equal(t, types.Duration("0"), pf.KeyTTL)

	require.Len(t, pf.Access, 2)
	// Access entries come from map iteration, so sort by name for stable assertion
	accessByName := make(map[string]types.AccessEntry)
	for _, a := range pf.Access {
		key := a.Name
		if key == "" {
			key = a.ID
		}
		accessByName[key] = a
	}
	usersEntry := accessByName["users-api"]
	assert.Equal(t, "users-api", usersEntry.Name)
	assert.Equal(t, []string{"v1"}, usersEntry.Versions)

	ordersEntry := accessByName["orders-api"]
	assert.Equal(t, "orders-api", ordersEntry.Name)
	assert.Equal(t, []string{"v1", "v2"}, ordersEntry.Versions)
}

// reqproof:req REQ-POL-007
func TestWireToCLI_FallbackToMID(t *testing.T) {
	dp := types.DashboardPolicy{
		MID:  "507f1f77bcf86cd799439011",
		ID:   "", // unmanaged policy — no tyk-cli: prefix
		Name: "Legacy Policy",
		AccessRights: map[string]*types.AccessRight{
			"a1b2c3d4e5f6": {
				APIID:   "a1b2c3d4e5f6",
				APIName: "users-api",
				Versions: []string{"v1"},
			},
		},
	}

	apis := []ResolverAPI{
		{ID: "a1b2c3d4e5f6", Name: "users-api"},
	}

	pf := WireToCLI(dp, apis)

	assert.Equal(t, "507f1f77bcf86cd799439011", pf.ID, "should fall back to MID when wire id is empty")
	assert.Equal(t, "Legacy Policy", pf.Name)
}

// reqproof:req REQ-POL-007
func TestRoundTrip_CLIToWireToCLI(t *testing.T) {
	// Original CLI policy
	original := types.PolicyFile{
		ID:   "silver",
		Name: "Silver Plan",
		Tags: []string{"silver"},
		RateLimit: &types.RateLimit{Requests: 500, Per: "1h"},
		Quota:     &types.Quota{Limit: 50000, Period: "1d"},
		KeyTTL:    "24h",
		Access: []types.AccessEntry{
			{Name: "users-api", Versions: []string{"v1"}},
		},
	}

	resolved := []ResolvedAccess{
		{APIID: "a1b2c3d4e5f6", APIName: "users-api", Versions: []string{"v1"}},
	}

	apis := []ResolverAPI{
		{ID: "a1b2c3d4e5f6", Name: "users-api"},
	}

	// CLI -> Wire
	wire, err := CLIToWire(original, resolved, "org-123")
	require.NoError(t, err)

	// Wire -> CLI
	roundTrip := WireToCLI(wire, apis)

	// Semantic equivalence (duration strings may normalize)
	assert.Equal(t, original.ID, roundTrip.ID)
	assert.Equal(t, original.Name, roundTrip.Name)
	assert.Equal(t, original.RateLimit.Requests, roundTrip.RateLimit.Requests)

	// Duration round-trip: "1h" -> 3600 -> "1h"
	assert.Equal(t, types.Duration("1h"), roundTrip.RateLimit.Per)
	assert.Equal(t, types.Duration("1d"), roundTrip.Quota.Period)
	assert.Equal(t, types.Duration("1d"), roundTrip.KeyTTL)

	require.Len(t, roundTrip.Access, 1)
	assert.Equal(t, "users-api", roundTrip.Access[0].Name)
	assert.Equal(t, []string{"v1"}, roundTrip.Access[0].Versions)
}

// ===========================================================================
// MC/DC coverage of CLIToWire branches (RateLimit/Quota/KeyTTL nil-or-set).
// ===========================================================================

// reqproof:req REQ-POL-007
func TestCLIToWire_BranchesMCDC(t *testing.T) {
	tests := []struct {
		name      string
		pf        types.PolicyFile
		wantErr   bool
		check     func(t *testing.T, dp types.DashboardPolicy)
	}{
		{
			name: "all duration fields nil/empty — none populated",
			pf: types.PolicyFile{
				ID:     "gold",
				Name:   "Gold",
				Access: []types.AccessEntry{{Name: "users-api", Versions: []string{"v1"}}},
			},
			check: func(t *testing.T, dp types.DashboardPolicy) {
				assert.EqualValues(t, 0, dp.Per)
				assert.EqualValues(t, 0, dp.QuotaRenewalRate)
				assert.EqualValues(t, 0, dp.KeyExpiresIn)
			},
		},
		{
			name: "rateLimit set but Per empty — Per stays 0",
			pf: types.PolicyFile{
				ID: "gold", Name: "Gold",
				RateLimit: &types.RateLimit{Requests: 100, Per: ""},
				Access:    []types.AccessEntry{{Name: "users-api", Versions: []string{"v1"}}},
			},
			check: func(t *testing.T, dp types.DashboardPolicy) {
				assert.EqualValues(t, 100, dp.Rate)
				assert.EqualValues(t, 0, dp.Per)
			},
		},
		{
			name: "rateLimit.Per invalid duration → error",
			pf: types.PolicyFile{
				ID: "gold", Name: "Gold",
				RateLimit: &types.RateLimit{Requests: 100, Per: "bogus"},
				Access:    []types.AccessEntry{{Name: "users-api", Versions: []string{"v1"}}},
			},
			wantErr: true,
		},
		{
			name: "quota set but Period empty",
			pf: types.PolicyFile{
				ID: "gold", Name: "Gold",
				Quota:  &types.Quota{Limit: 1000, Period: ""},
				Access: []types.AccessEntry{{Name: "users-api", Versions: []string{"v1"}}},
			},
			check: func(t *testing.T, dp types.DashboardPolicy) {
				assert.EqualValues(t, 1000, dp.QuotaMax)
				assert.EqualValues(t, 0, dp.QuotaRenewalRate)
			},
		},
		{
			name: "quota.Period invalid duration → error",
			pf: types.PolicyFile{
				ID: "gold", Name: "Gold",
				Quota:  &types.Quota{Limit: 1000, Period: "bogus"},
				Access: []types.AccessEntry{{Name: "users-api", Versions: []string{"v1"}}},
			},
			wantErr: true,
		},
		{
			name: "keyTTL invalid duration → error",
			pf: types.PolicyFile{
				ID: "gold", Name: "Gold",
				KeyTTL: "bogus",
				Access: []types.AccessEntry{{Name: "users-api", Versions: []string{"v1"}}},
			},
			wantErr: true,
		},
		{
			name: "all fields populated and valid",
			pf: types.PolicyFile{
				ID:        "gold",
				Name:      "Gold",
				RateLimit: &types.RateLimit{Requests: 100, Per: "60"},
				Quota:     &types.Quota{Limit: 1000, Period: "30d"},
				KeyTTL:    "1h",
				Access:    []types.AccessEntry{{Name: "users-api", Versions: []string{"v1"}}},
			},
			check: func(t *testing.T, dp types.DashboardPolicy) {
				assert.EqualValues(t, 60, dp.Per)
				assert.EqualValues(t, 2592000, dp.QuotaRenewalRate)
				assert.EqualValues(t, 3600, dp.KeyExpiresIn)
			},
		},
	}

	apis := []ResolverAPI{
		{ID: "a1b2c3d4e5f6", Name: "users-api"},
	}
	resolveReqs := []ResolveRequest{
		{SelectorType: "name", Value: "users-api", Versions: []string{"v1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, _ := ResolveAccessEntries(resolveReqs, apis)
			dp, err := CLIToWire(tt.pf, resolved, "test-org")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.check != nil {
				tt.check(t, dp)
			}
		})
	}
}

// ===========================================================================
// MC/DC coverage of WireToCLI branches.
// ===========================================================================

// reqproof:req REQ-POL-007
func TestWireToCLI_BranchesMCDC(t *testing.T) {
	apis := []ResolverAPI{
		{ID: "a1b2c3d4e5f6", Name: "users-api"},
	}

	t.Run("MID fallback when ID empty", func(t *testing.T) {
		dp := types.DashboardPolicy{
			MID: "507f1f77bcf86cd799439020", ID: "",
			Name: "Free", AccessRights: map[string]*types.AccessRight{},
		}
		pf := WireToCLI(dp, apis)
		assert.Equal(t, "507f1f77bcf86cd799439020", pf.ID, "MID must be used when ID is empty")
	})

	t.Run("ID wins over MID when both set", func(t *testing.T) {
		dp := types.DashboardPolicy{
			MID: "507f1f77bcf86cd799439020", ID: "free-tier",
			Name: "Free", AccessRights: map[string]*types.AccessRight{},
		}
		pf := WireToCLI(dp, apis)
		assert.Equal(t, "free-tier", pf.ID, "friendly ID must take precedence")
	})

	t.Run("no rate-limit when both Rate and Per are zero", func(t *testing.T) {
		dp := types.DashboardPolicy{
			ID: "x", Name: "x", Rate: 0, Per: 0,
			AccessRights: map[string]*types.AccessRight{},
		}
		pf := WireToCLI(dp, apis)
		assert.Nil(t, pf.RateLimit, "RateLimit must be nil when both fields are zero")
	})

	t.Run("rate-limit set when only Per > 0", func(t *testing.T) {
		dp := types.DashboardPolicy{
			ID: "x", Name: "x", Rate: 0, Per: 60,
			AccessRights: map[string]*types.AccessRight{},
		}
		pf := WireToCLI(dp, apis)
		assert.NotNil(t, pf.RateLimit, "RateLimit must be populated when Per > 0 even if Rate is 0")
	})

	t.Run("no quota when both fields are zero", func(t *testing.T) {
		dp := types.DashboardPolicy{
			ID: "x", Name: "x", QuotaMax: 0, QuotaRenewalRate: 0,
			AccessRights: map[string]*types.AccessRight{},
		}
		pf := WireToCLI(dp, apis)
		assert.Nil(t, pf.Quota, "Quota must be nil when both fields are zero")
	})

	// MC/DC: prove QuotaRenewalRate > 0 independently of QuotaMax > 0 by
	// providing only the renewal-rate field (convert.go:92 short-circuit gap).
	t.Run("quota set when only QuotaRenewalRate > 0", func(t *testing.T) {
		dp := types.DashboardPolicy{
			ID: "x", Name: "x", QuotaMax: 0, QuotaRenewalRate: 86400,
			AccessRights: map[string]*types.AccessRight{},
		}
		pf := WireToCLI(dp, apis)
		require.NotNil(t, pf.Quota, "Quota must be populated when QuotaRenewalRate > 0 even if QuotaMax is 0")
		assert.EqualValues(t, 0, pf.Quota.Limit)
		assert.Equal(t, types.Duration("1d"), pf.Quota.Period)
	})

	t.Run("access rights with unknown API id falls back to ID", func(t *testing.T) {
		dp := types.DashboardPolicy{
			ID: "x", Name: "x",
			AccessRights: map[string]*types.AccessRight{
				"unknown-id-1234": {APIID: "unknown-id-1234", Versions: []string{"v1"}},
			},
		}
		pf := WireToCLI(dp, apis)
		require := len(pf.Access) == 1
		assert.True(t, require, "expected exactly one access entry")
		if require {
			assert.Equal(t, "unknown-id-1234", pf.Access[0].ID, "unknown API should fall back to ID selector")
			assert.Empty(t, pf.Access[0].Name, "no Name should be set when API is unknown")
		}
	})

	t.Run("access rights with known API uses name selector", func(t *testing.T) {
		dp := types.DashboardPolicy{
			ID: "x", Name: "x",
			AccessRights: map[string]*types.AccessRight{
				"a1b2c3d4e5f6": {APIID: "a1b2c3d4e5f6", Versions: []string{"v1"}},
			},
		}
		pf := WireToCLI(dp, apis)
		require := len(pf.Access) == 1
		assert.True(t, require, "expected exactly one access entry")
		if require {
			assert.Equal(t, "users-api", pf.Access[0].Name, "known API should be referenced by name")
			assert.Empty(t, pf.Access[0].ID, "no ID should be set when name is available")
		}
	})
}
