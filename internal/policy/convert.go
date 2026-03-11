package policy

import (
	"fmt"
	"sort"

	"github.com/tyktech/tyk-cli/pkg/types"
)

// CLIToWire converts a PolicyFile and pre-resolved access entries into the Dashboard wire format.
// The caller is responsible for resolving selectors before calling this function.
func CLIToWire(pf types.PolicyFile, resolved []ResolvedAccess, orgID string) (types.DashboardPolicy, error) {
	dp := types.DashboardPolicy{
		// MID intentionally left empty — caller sets it after resolution
		ID:         pf.ID,
		Name:       pf.Name,
		OrgID:      orgID,
		Tags:       pf.Tags,
		Active:     true,
		IsInactive: false,
	}

	// Rate limit
	if pf.RateLimit != nil {
		dp.Rate = pf.RateLimit.Requests
		if pf.RateLimit.Per != "" {
			per, err := ParseDuration(string(pf.RateLimit.Per))
			if err != nil {
				return types.DashboardPolicy{}, fmt.Errorf("rateLimit.per: %w", err)
			}
			dp.Per = per
		}
	}

	// Quota
	if pf.Quota != nil {
		dp.QuotaMax = pf.Quota.Limit
		if pf.Quota.Period != "" {
			period, err := ParseDuration(string(pf.Quota.Period))
			if err != nil {
				return types.DashboardPolicy{}, fmt.Errorf("quota.period: %w", err)
			}
			dp.QuotaRenewalRate = period
		}
	}

	// Key TTL
	if pf.KeyTTL != "" {
		ttl, err := ParseDuration(string(pf.KeyTTL))
		if err != nil {
			return types.DashboardPolicy{}, fmt.Errorf("keyTTL: %w", err)
		}
		dp.KeyExpiresIn = ttl
	}

	// Access rights from resolved entries
	dp.AccessRights = make(map[string]*types.AccessRight, len(resolved))
	for _, r := range resolved {
		dp.AccessRights[r.APIID] = &types.AccessRight{
			APIID:       r.APIID,
			APIName:     r.APIName,
			Versions:    r.Versions,
			AllowedURLs: nil,
			Limit:       nil,
		}
	}

	return dp, nil
}

// WireToCLI converts a DashboardPolicy back to the CLI PolicyFile format.
// It uses the provided API list for best-effort reverse resolution of API IDs to names.
func WireToCLI(dp types.DashboardPolicy) types.PolicyFile {
	friendlyID := dp.ID
	if friendlyID == "" {
		friendlyID = dp.MID // fallback for unmanaged policies
	}

	pf := types.PolicyFile{
		ID:   friendlyID,
		Name: dp.Name,
		Tags: dp.Tags,
	}

	// Rate limit
	if dp.Rate > 0 || dp.Per > 0 {
		pf.RateLimit = &types.RateLimit{
			Requests: dp.Rate,
			Per:      types.Duration(FormatDuration(dp.Per)),
		}
	}

	// Quota
	if dp.QuotaMax > 0 || dp.QuotaRenewalRate > 0 {
		pf.Quota = &types.Quota{
			Limit:  dp.QuotaMax,
			Period: types.Duration(FormatDuration(dp.QuotaRenewalRate)),
		}
	}

	// Key TTL
	pf.KeyTTL = types.Duration(FormatDuration(dp.KeyExpiresIn))

	// Access rights -> access entries, sorted by API ID for deterministic output.
	// The wire format already contains api_name, so no external API list lookup is needed.
	apiIDs := make([]string, 0, len(dp.AccessRights))
	for id := range dp.AccessRights {
		apiIDs = append(apiIDs, id)
	}
	sort.Strings(apiIDs)

	for _, apiID := range apiIDs {
		ar := dp.AccessRights[apiID]
		entry := types.AccessEntry{
			Versions: ar.Versions,
		}

		// Use the name from the wire format; fall back to raw ID if empty
		if ar.APIName != "" {
			entry.Name = ar.APIName
		} else {
			entry.ID = apiID
		}

		pf.Access = append(pf.Access, entry)
	}

	return pf
}
