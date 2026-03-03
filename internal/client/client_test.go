package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Helper function to create a valid config with environment
func createTestConfig(dashboardURL, authToken, orgID string) *types.Config {
	return &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {
				Name:         "test",
				DashboardURL: dashboardURL,
				AuthToken:    authToken,
				OrgID:        orgID,
			},
		},
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      createTestConfig("http://localhost:3000", "test-token", "test-org"),
			expectError: false,
		},
		{
			name: "invalid config - no environments",
			config: &types.Config{
				DefaultEnvironment: "",
				Environments:       make(map[string]*types.Environment),
			},
			expectError: true,
		},
		{
			name:        "invalid URL format",
			config:      createTestConfig("invalid-url", "test-token", "test-org"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.config, client.config)
			}
		})
	}
}

func TestClient_SetTimeout(t *testing.T) {
	config := createTestConfig("http://localhost:3000", "test-token", "test-org")

	client, err := NewClient(config)
	require.NoError(t, err)

	newTimeout := 45 * time.Second
	client.SetTimeout(newTimeout)
	assert.Equal(t, newTimeout, client.httpClient.Timeout)
}

func TestClient_doRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.Equal(t, "test-token", r.Header.Get("authorization"))
		assert.Equal(t, "application/json", r.Header.Get("accept"))

		// Echo back request info
		response := map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": "success",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")

	client, err := NewClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	resp, err := client.doRequest(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_handleResponse(t *testing.T) {
	config := createTestConfig("http://localhost:3000", "test-token", "test-org")

	client, err := NewClient(config)
	require.NoError(t, err)

	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]string{"status": "success", "message": "OK"}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)

		var result map[string]string
		err = client.handleResponse(resp, &result)
		assert.NoError(t, err)
		assert.Equal(t, "success", result["status"])
		assert.Equal(t, "OK", result["message"])
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			response := map[string]interface{}{
				"status":  404,
				"message": "API not found",
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		resp, err := http.Get(server.URL)
		require.NoError(t, err)

		err = client.handleResponse(resp, nil)
		assert.Error(t, err)

		errorResp, ok := err.(*types.ErrorResponse)
		assert.True(t, ok)
		assert.Equal(t, 404, errorResp.Status)
		assert.Contains(t, errorResp.Message, "API not found")
	})
}

func TestClient_GetOASAPI(t *testing.T) {
	// Create a mock OAS document with x-tyk-api-gateway extension
	mockOASDoc := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Test API",
			"version": "1.0.0",
		},
		"x-tyk-api-gateway": map[string]interface{}{
			"info": map[string]interface{}{
				"id":   "test-api-id",
				"name": "Test API",
			},
			"server": map[string]interface{}{
				"listenPath": map[string]interface{}{
					"value": "/test",
				},
			},
			"upstream": map[string]interface{}{
				"url": "http://example.com",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/apis/oas/test-api-id")

		// Return raw OAS document (as the Tyk Dashboard does)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOASDoc)
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")

	client, err := NewClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	api, err := client.GetOASAPI(ctx, "test-api-id", "")
	require.NoError(t, err)
	assert.Equal(t, "test-api-id", api.ID)
	assert.Equal(t, "Test API", api.Name)
	assert.Equal(t, "/test", api.ListenPath)
	assert.Equal(t, "http://example.com", api.UpstreamURL)
}

func TestClient_CreateOASAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/apis/oas" {
			// Handle create request - return basic response with ID
			response := types.APIResponse{Status: "success", ID: "new-api-id"}
			_ = json.NewEncoder(w).Encode(response)
		} else if r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/new-api-id" {
			// Return raw OAS document similar to Dashboard
			oasDoc := map[string]interface{}{
				"openapi": "3.0.0",
				"info": map[string]interface{}{
					"title":   "New API",
					"version": "1.0.0",
				},
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{
						"id":   "new-api-id",
						"name": "New API",
					},
					"server": map[string]interface{}{
						"listenPath": map[string]interface{}{
							"value": "/new",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(oasDoc)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")

	client, err := NewClient(config)
	require.NoError(t, err)

	oasDoc := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Test API",
			"version": "1.0.0",
		},
	}

	ctx := context.Background()
	api, err := client.CreateOASAPI(ctx, oasDoc)
	require.NoError(t, err)
	assert.Equal(t, "new-api-id", api.ID)
	assert.Equal(t, "New API", api.Name)
}

func TestClient_ListOASAPIs(t *testing.T) {
	// Prepare two mock APIs
	mockAPIs := []*types.OASAPI{
		{ID: "api-1", Name: "API One", ListenPath: "/one", DefaultVersion: "v1"},
		{ID: "api-2", Name: "API Two", ListenPath: "/two", DefaultVersion: "v1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/apis/oas", r.URL.Path)
		// Ensure pagination param is passed when provided
		assert.Equal(t, "2", r.URL.Query().Get("p"))
		_ = json.NewEncoder(w).Encode(types.OASAPIListResponse{
			APIResponse: types.APIResponse{Status: "success"},
			APIs:        mockAPIs,
		})
	}))
	defer server.Close()

	config := createTestConfig(server.URL, "test-token", "test-org")

	client, err := NewClient(config)
	require.NoError(t, err)

	ctx := context.Background()
	apis, err := client.ListOASAPIs(ctx, 2)
	require.NoError(t, err)
	require.Len(t, apis, 2)
	assert.Equal(t, "api-1", apis[0].ID)
	assert.Equal(t, "API One", apis[0].Name)
}

func TestClient_Health(t *testing.T) {
	t.Run("healthy dashboard", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/health", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}))
		defer server.Close()

		config := createTestConfig(server.URL, "test-token", "test-org")

		client, err := NewClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = client.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("unhealthy dashboard", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Service Unavailable"))
		}))
		defer server.Close()

		config := createTestConfig(server.URL, "test-token", "test-org")

		client, err := NewClient(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = client.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed")
	})
}

func TestClient_CreateOASAPIVersion(t *testing.T) {
	t.Run("success posts to correct endpoint with query params and body", func(t *testing.T) {
		oasBody := json.RawMessage(`{"openapi":"3.0.0","info":{"title":"V2","version":"2.0.0"}}`)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/apis/oas", r.URL.Path)
			assert.Equal(t, "base-123", r.URL.Query().Get("base_api_id"))
			assert.Equal(t, "v2", r.URL.Query().Get("new_version_name"))
			assert.Equal(t, "true", r.URL.Query().Get("set_default"))

			// Verify body is the OAS doc
			var body map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "3.0.0", body["openapi"])

			resp := types.APIResponse{Status: "success", Message: "created", ID: "new-version-id"}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		config := createTestConfig(server.URL, "test-token", "test-org")
		client, err := NewClient(config)
		require.NoError(t, err)

		result, err := client.CreateOASAPIVersion(context.Background(), oasBody, "base-123", "v2", true)
		require.NoError(t, err)
		assert.Equal(t, "success", result.Status)
		assert.Equal(t, "new-version-id", result.ID)
	})

	t.Run("passes through dashboard error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]interface{}{"status": 400, "message": "base API not found"}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		config := createTestConfig(server.URL, "test-token", "test-org")
		client, err := NewClient(config)
		require.NoError(t, err)

		result, err := client.CreateOASAPIVersion(context.Background(), json.RawMessage(`{}`), "bad-id", "v2", false)
		assert.Error(t, err)
		assert.Nil(t, result)

		errorResp, ok := err.(*types.ErrorResponse)
		assert.True(t, ok)
		assert.Equal(t, 400, errorResp.Status)
	})
}

// Integration test with live environment
func TestLiveEnvironmentClient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	config := createTestConfig("http://tyk-dashboard.localhost:3000", "ff8289874f5d45de945a2ea5c02580fe", "5e9d9544a1dcd60001d0ed20")

	client, err := NewClient(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("health check", func(t *testing.T) {
		err := client.Health(ctx)
		if err != nil {
			t.Logf("Live environment health check failed: %v", err)
			t.Skip("Live environment not available, skipping integration test")
		}
		t.Log("✓ Live environment health check passed")
	})

	// Only run API tests if health check passed
	if client.Health(ctx) == nil {
		t.Run("test API endpoint", func(t *testing.T) {
			_, err := client.GetOASAPI(ctx, "non-existent-api-id", "")
			assert.Error(t, err)

			// Check that it's a proper error response
			errorResp, ok := err.(*types.ErrorResponse)
			if ok {
				t.Logf("✓ Received proper error response: status=%d, message=%s", errorResp.Status, errorResp.Message)
				// Accept any error status (401, 404, etc.) as proof the API is working
				assert.True(t, errorResp.Status >= 400)
			}
		})
	}
}
