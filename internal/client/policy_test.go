package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------

// reqproof:req REQ-POL-001
func sampleDashboardPolicy(id, name string, rate int64) types.DashboardPolicy {
	return types.DashboardPolicy{
		MID:              id,
		ID:               "",
		Name:             name,
		OrgID:            "test-org",
		Rate:             rate,
		Per:              60,
		QuotaMax:         100000,
		QuotaRenewalRate: 2592000,
		Tags:             []string{"test"},
		AccessRights:     map[string]*types.AccessRight{},
		Active:           true,
		IsInactive:       false,
	}
}

// ---------------------------------------------------------------------------
// ListPolicies
// ---------------------------------------------------------------------------

// reqproof:req REQ-POL-001
func TestClient_ListPolicies_ZeroPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// page <= 0 must NOT append the `p=` query parameter.
		assert.Empty(t, r.URL.RawQuery)
		_ = json.NewEncoder(w).Encode(types.DashboardPolicyListResponse{Data: nil, Pages: 0})
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	result, err := client.ListPolicies(context.Background(), 0)
	require.NoError(t, err)
	assert.Empty(t, result.Data)
}

// reqproof:req REQ-POL-001
func TestClient_ListPolicies_HandleResponseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"status":500,"message":"server error"}`))
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	_, err = client.ListPolicies(context.Background(), 1)
	require.Error(t, err)
	errResp, ok := err.(*types.ErrorResponse)
	require.True(t, ok)
	assert.Equal(t, 500, errResp.Status)
}

// reqproof:req REQ-POL-001
func TestClient_ListPolicies_NetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedURL := server.URL
	server.Close()

	config := createTestConfig(closedURL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)
	client.SetTimeout(200 * time.Millisecond)

	_, err = client.ListPolicies(context.Background(), 0)
	require.Error(t, err)
}

// reqproof:req REQ-POL-001
func TestClient_ListPolicies(t *testing.T) {
	gold := sampleDashboardPolicy("gold", "Gold Plan", 1000)
	silver := sampleDashboardPolicy("silver", "Silver Plan", 500)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/portal/policies", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("p"))
		assert.Equal(t, "test-token", r.Header.Get("authorization"))

		resp := types.DashboardPolicyListResponse{
			Data:  []types.DashboardPolicy{gold, silver},
			Pages: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	result, err := client.ListPolicies(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, result.Data, 2)
	assert.Equal(t, "gold", result.Data[0].MID)
	assert.Equal(t, "Gold Plan", result.Data[0].Name)
	assert.Equal(t, "silver", result.Data[1].MID)
}

// reqproof:req REQ-POL-001
func TestClient_ListPolicies_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.DashboardPolicyListResponse{
			Data:  nil,
			Pages: 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	result, err := client.ListPolicies(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, result.Data)
}

// ---------------------------------------------------------------------------
// GetPolicy
// ---------------------------------------------------------------------------

// reqproof:req REQ-POL-002
func TestClient_GetPolicy(t *testing.T) {
	gold := sampleDashboardPolicy("gold", "Gold Plan", 1000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/portal/policies/gold", r.URL.Path)
		_ = json.NewEncoder(w).Encode(gold)
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	result, err := client.GetPolicy(context.Background(), "gold")
	require.NoError(t, err)
	assert.Equal(t, "gold", result.MID)
	assert.Equal(t, "Gold Plan", result.Name)
	assert.Equal(t, int64(1000), result.Rate)
}

// reqproof:req REQ-POL-022
func TestClient_GetPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404, "message": "policy not found",
		})
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	_, err = client.GetPolicy(context.Background(), "nonexistent")
	require.Error(t, err)
	errorResp, ok := err.(*types.ErrorResponse)
	require.True(t, ok, "expected *types.ErrorResponse, got %T", err)
	assert.Equal(t, 404, errorResp.Status)
	assert.Contains(t, errorResp.Message, "policy not found")
}

// ---------------------------------------------------------------------------
// CreatePolicy
// ---------------------------------------------------------------------------

// reqproof:req REQ-POL-003
func TestClient_CreatePolicy(t *testing.T) {
	policy := sampleDashboardPolicy("new-policy", "New Policy", 500)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/portal/policies", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		body, _ := io.ReadAll(r.Body)
		var payload types.DashboardPolicy
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "new-policy", payload.MID)

		// Dashboard returns a Message response with the created policy ID in Meta
		_ = json.NewEncoder(w).Encode(types.APIResponse{
			Status:  "success",
			Message: "created",
			Meta:    "new-policy",
		})
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	err = client.CreatePolicy(context.Background(), &policy)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// UpdatePolicy
// ---------------------------------------------------------------------------

// reqproof:req REQ-POL-003
func TestClient_UpdatePolicy(t *testing.T) {
	policy := sampleDashboardPolicy("gold", "Gold Plan Updated", 2000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/portal/policies/gold", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("content-type"))

		body, _ := io.ReadAll(r.Body)
		var payload types.DashboardPolicy
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, "Gold Plan Updated", payload.Name)

		_ = json.NewEncoder(w).Encode(types.APIResponse{
			Status:  "success",
			Message: "updated",
		})
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	err = client.UpdatePolicy(context.Background(), "gold", &policy)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// DeletePolicy
// ---------------------------------------------------------------------------

// reqproof:req REQ-POL-004
func TestClient_DeletePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/portal/policies/free-tier", r.URL.Path)

		_ = json.NewEncoder(w).Encode(types.APIResponse{
			Status:  "success",
			Message: "deleted",
		})
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	err = client.DeletePolicy(context.Background(), "free-tier")
	require.NoError(t, err)
}

// reqproof:req REQ-POL-022
func TestClient_DeletePolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404, "message": "policy not found",
		})
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)

	err = client.DeletePolicy(context.Background(), "nonexistent")
	require.Error(t, err)
	errorResp, ok := err.(*types.ErrorResponse)
	require.True(t, ok, "expected *types.ErrorResponse, got %T", err)
	assert.Equal(t, 404, errorResp.Status)
}

// ---------------------------------------------------------------------------
// doRequest network-failure coverage for every policy verb.
// ---------------------------------------------------------------------------

// reqproof:req REQ-POL-002
func TestClient_GetPolicy_NetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedURL := server.URL
	server.Close()

	config := createTestConfig(closedURL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)
	client.SetTimeout(200 * time.Millisecond)

	_, err = client.GetPolicy(context.Background(), "x")
	require.Error(t, err)
}

// reqproof:req REQ-POL-003
func TestClient_CreatePolicy_NetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedURL := server.URL
	server.Close()

	config := createTestConfig(closedURL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)
	client.SetTimeout(200 * time.Millisecond)

	policy := sampleDashboardPolicy("x", "X", 1)
	err = client.CreatePolicy(context.Background(), &policy)
	require.Error(t, err)
}

// reqproof:req REQ-POL-003
func TestClient_UpdatePolicy_NetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedURL := server.URL
	server.Close()

	config := createTestConfig(closedURL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)
	client.SetTimeout(200 * time.Millisecond)

	policy := sampleDashboardPolicy("x", "X", 1)
	err = client.UpdatePolicy(context.Background(), "x", &policy)
	require.Error(t, err)
}

// reqproof:req REQ-POL-004
func TestClient_DeletePolicy_NetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedURL := server.URL
	server.Close()

	config := createTestConfig(closedURL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)
	client.SetTimeout(200 * time.Millisecond)

	err = client.DeletePolicy(context.Background(), "x")
	require.Error(t, err)
}
