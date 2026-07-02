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

// Implements: SYS-REQ-024
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

// Implements: SYS-REQ-025
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

// Implements: SYS-REQ-026
func (c *Client) CreatePolicy(ctx context.Context, policy *types.DashboardPolicy) error {
	resp, err := c.doRequest(ctx, http.MethodPost, PoliciesPath, policy)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

// Implements: SYS-REQ-026
func (c *Client) UpdatePolicy(ctx context.Context, policyID string, policy *types.DashboardPolicy) error {
	policyPath := fmt.Sprintf(PolicyPath, url.PathEscape(policyID))

	resp, err := c.doRequest(ctx, http.MethodPut, policyPath, policy)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}

// Implements: SYS-REQ-027
func (c *Client) DeletePolicy(ctx context.Context, policyID string) error {
	policyPath := fmt.Sprintf(PolicyPath, url.PathEscape(policyID))

	resp, err := c.doRequest(ctx, http.MethodDelete, policyPath, nil)
	if err != nil {
		return err
	}

	return c.handleResponse(resp, nil)
}
