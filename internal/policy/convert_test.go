package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

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

	assert.Equal(t, "gold", dp.MID)
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

func TestWireToCLI(t *testing.T) {
	dp := types.DashboardPolicy{
		MID:              "gold",
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

	assert.Equal(t, "gold", pf.ID)
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
