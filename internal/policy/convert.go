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
		MID:        pf.Metadata.ID,
		Name:       pf.Metadata.Name,
		OrgID:      orgID,
		Tags:       pf.Metadata.Tags,
		Active:     true,
		IsInactive: false,
	}

	// Rate limit
	if pf.Spec.RateLimit != nil {
		dp.Rate = pf.Spec.RateLimit.Requests
		if pf.Spec.RateLimit.Per != "" {
			per, err := ParseDuration(string(pf.Spec.RateLimit.Per))
			if err != nil {
				return types.DashboardPolicy{}, fmt.Errorf("rateLimit.per: %w", err)
			}
			dp.Per = per
		}
	}

	// Quota
	if pf.Spec.Quota != nil {
		dp.QuotaMax = pf.Spec.Quota.Limit
		if pf.Spec.Quota.Period != "" {
			period, err := ParseDuration(string(pf.Spec.Quota.Period))
			if err != nil {
				return types.DashboardPolicy{}, fmt.Errorf("quota.period: %w", err)
			}
			dp.QuotaRenewalRate = period
		}
	}

	// Key TTL
	if pf.Spec.KeyTTL != "" {
		ttl, err := ParseDuration(string(pf.Spec.KeyTTL))
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
func WireToCLI(dp types.DashboardPolicy, apis []ResolverAPI) types.PolicyFile {
	pf := types.PolicyFile{
		APIVersion: "tyk.tyktech/v1",
		Kind:       "Policy",
		Metadata: types.PolicyMetadata{
			ID:   dp.MID,
			Name: dp.Name,
			Tags: dp.Tags,
		},
	}

	// Rate limit
	if dp.Rate > 0 || dp.Per > 0 {
		pf.Spec.RateLimit = &types.RateLimit{
			Requests: dp.Rate,
			Per:      types.Duration(FormatDuration(dp.Per)),
		}
	}

	// Quota
	if dp.QuotaMax > 0 || dp.QuotaRenewalRate > 0 {
		pf.Spec.Quota = &types.Quota{
			Limit:  dp.QuotaMax,
			Period: types.Duration(FormatDuration(dp.QuotaRenewalRate)),
		}
	}

	// Key TTL
	pf.Spec.KeyTTL = types.Duration(FormatDuration(dp.KeyExpiresIn))

	// Build API name lookup
	apiByID := make(map[string]ResolverAPI, len(apis))
	for _, api := range apis {
		apiByID[api.ID] = api
	}

	// Access rights -> access entries, sorted by API ID for deterministic output
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

		// Best-effort reverse resolution: use name if API is known, otherwise fall back to ID
		if api, ok := apiByID[apiID]; ok {
			entry.Name = api.Name
		} else {
			entry.ID = apiID
		}

		pf.Spec.Access = append(pf.Spec.Access, entry)
	}

	return pf
}
