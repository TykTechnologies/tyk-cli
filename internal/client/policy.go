package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tyktech/tyk-cli/pkg/types"
)

const (
	// Policy endpoints
	PoliciesPath = "/api/portal/policies"
	PolicyPath   = "/api/portal/policies/%s" // {policyId}
)

// ListPolicies retrieves a paginated list of policies from the Dashboard.
// Page numbers are 1-based.
func (c *Client) ListPolicies(ctx context.Context, page int) (*types.DashboardPolicyListResponse, error) {
	listPath := PoliciesPath
	if page > 0 {
		values := url.Values{}
		values.Set("p", fmt.Sprintf("%d", page))
		listPath += "?" + values.Encode()
	}

	resp, err := c.doRequest(ctx, http.MethodGet, listPath, nil)
	if err != nil {
		return nil, err
	}

	var result types.DashboardPolicyListResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetPolicy retrieves a single policy by ID.
// Returns *types.ErrorResponse on 404.
func (c *Client) GetPolicy(ctx context.Context, policyID string) (*types.DashboardPolicy, error) {
	policyPath := fmt.Sprintf(PolicyPath, url.PathEscape(policyID))

	resp, err := c.doRequest(ctx, http.MethodGet, policyPath, nil)
	if err != nil {
		return nil, err
	}

	var result types.DashboardPolicy
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreatePolicy sends a POST request with a DashboardPolicy JSON body.
func (c *Client) CreatePolicy(ctx context.Context, policy *types.DashboardPolicy) error {
	resp, err := c.doRequest(ctx, http.MethodPost, PoliciesPath, policy)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

// UpdatePolicy sends a PUT request to update an existing policy by ID.
func (c *Client) UpdatePolicy(ctx context.Context, policyID string, policy *types.DashboardPolicy) error {
	policyPath := fmt.Sprintf(PolicyPath, url.PathEscape(policyID))

	resp, err := c.doRequest(ctx, http.MethodPut, policyPath, policy)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

// DeletePolicy sends a DELETE request to remove a policy by ID.
// Returns *types.ErrorResponse on 404.
func (c *Client) DeletePolicy(ctx context.Context, policyID string) error {
	policyPath := fmt.Sprintf(PolicyPath, url.PathEscape(policyID))

	resp, err := c.doRequest(ctx, http.MethodDelete, policyPath, nil)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}
