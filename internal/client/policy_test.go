package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------

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
		json.NewEncoder(w).Encode(resp)
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

func TestClient_ListPolicies_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := types.DashboardPolicyListResponse{
			Data:  nil,
			Pages: 0,
		}
		json.NewEncoder(w).Encode(resp)
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

func TestClient_GetPolicy(t *testing.T) {
	gold := sampleDashboardPolicy("gold", "Gold Plan", 1000)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/portal/policies/gold", r.URL.Path)
		json.NewEncoder(w).Encode(gold)
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

func TestClient_GetPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(types.APIResponse{
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

		json.NewEncoder(w).Encode(types.APIResponse{
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

func TestClient_DeletePolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/portal/policies/free-tier", r.URL.Path)

		json.NewEncoder(w).Encode(types.APIResponse{
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

func TestClient_DeletePolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
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
