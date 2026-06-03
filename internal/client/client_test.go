package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// reqproof:req REQ-API-030
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

// reqproof:req REQ-API-030
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

// reqproof:req REQ-API-030
func TestClient_SetTimeout(t *testing.T) {
	config := createTestConfig("http://localhost:3000", "test-token", "test-org")

	client, err := NewClient(config)
	require.NoError(t, err)

	newTimeout := 45 * time.Second
	client.SetTimeout(newTimeout)
	assert.Equal(t, newTimeout, client.httpClient.Timeout)
}

// reqproof:req REQ-CFG-007
func TestClient_HonoursEnvironmentTimeout(t *testing.T) {
	t.Run("environment timeout overrides default", func(t *testing.T) {
		config := createTestConfig("http://localhost:3000", "test-token", "test-org")
		config.Environments["test"].TimeoutSeconds = 7

		client, err := NewClient(config)
		require.NoError(t, err)
		assert.Equal(t, 7*time.Second, client.httpClient.Timeout,
			"REQ-CFG-007: env-level timeout_seconds must be applied to http.Client.Timeout")
	})

	t.Run("zero environment timeout falls back to DefaultTimeout", func(t *testing.T) {
		config := createTestConfig("http://localhost:3000", "test-token", "test-org")
		config.Environments["test"].TimeoutSeconds = 0

		client, err := NewClient(config)
		require.NoError(t, err)
		assert.Equal(t, DefaultTimeout, client.httpClient.Timeout,
			"zero timeout_seconds must fall through to the 30s default")
	})

	t.Run("negative environment timeout falls back to DefaultTimeout", func(t *testing.T) {
		config := createTestConfig("http://localhost:3000", "test-token", "test-org")
		config.Environments["test"].TimeoutSeconds = -5

		client, err := NewClient(config)
		require.NoError(t, err)
		assert.Equal(t, DefaultTimeout, client.httpClient.Timeout,
			"negative timeout_seconds must fall through to the 30s default")
	})
}

// reqproof:req REQ-API-030
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

// reqproof:req REQ-API-020
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

// reqproof:req REQ-API-002
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

// reqproof:req REQ-API-003
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

// reqproof:req REQ-API-001
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

// reqproof:req REQ-CFG-002
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

// ---------------------------------------------------------------------------
// doRequest error-path coverage
// ---------------------------------------------------------------------------

// newClientForTest constructs a Client pointed at the provided URL and returns it.
// reqproof:req REQ-API-030
func newClientForTest(t *testing.T, dashboardURL string) *Client {
	t.Helper()
	config := createTestConfig(dashboardURL, "test-token", "test-org")
	client, err := NewClient(config)
	require.NoError(t, err)
	return client
}

// reqproof:req REQ-API-030
func TestClient_doRequest_ErrorPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("json marshal failure for unsupported body type", func(t *testing.T) {
		// A channel cannot be marshalled to JSON, so json.Marshal returns an error
		// and doRequest must surface it before any HTTP traffic happens.
		client := newClientForTest(t, "http://127.0.0.1:0")
		ch := make(chan int)
		resp, err := client.doRequest(ctx, http.MethodPost, "/x", ch)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to marshal request body")
	})

	t.Run("invalid http method causes NewRequestWithContext failure", func(t *testing.T) {
		client := newClientForTest(t, "http://127.0.0.1:0")
		resp, err := client.doRequest(ctx, "BAD METHOD", "/x", nil)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to create request")
	})

	t.Run("active environment becomes invalid after construction", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		// Mutate the underlying config so GetActiveEnvironment fails on the
		// auth-lookup branch in doRequest.
		client.config.DefaultEnvironment = ""
		client.config.Environments = map[string]*types.Environment{}

		resp, err := client.doRequest(ctx, http.MethodGet, "/x", nil)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no active environment for auth")
	})

	t.Run("network failure when server is unreachable", func(t *testing.T) {
		// Start then stop a server to guarantee a closed port for the dial.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.doRequest(ctx, http.MethodGet, "/x", nil)
		require.Error(t, err)
	})

	t.Run("string body sets json content type and reaches server", func(t *testing.T) {
		var gotCT string
		var gotBody string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotCT = r.Header.Get("content-type")
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		resp, err := client.doRequest(ctx, http.MethodPost, "/x", `{"k":"v"}`)
		require.NoError(t, err)
		require.NotNil(t, resp)
		_ = resp.Body.Close()
		assert.Equal(t, "application/json", gotCT)
		assert.Equal(t, `{"k":"v"}`, gotBody)
	})

	t.Run("byte-slice body sets json content type and reaches server", func(t *testing.T) {
		var gotBody string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		resp, err := client.doRequest(ctx, http.MethodPost, "/x", []byte(`{"a":1}`))
		require.NoError(t, err)
		require.NotNil(t, resp)
		_ = resp.Body.Close()
		assert.Equal(t, `{"a":1}`, gotBody)
	})

	t.Run("path without query string falls into else branch", func(t *testing.T) {
		// "%zz" is an invalid percent escape so url.Parse(path) returns an
		// error and doRequest falls into the else branch that assigns the raw
		// path. The downstream http.NewRequestWithContext then returns an
		// error because the URL is unparseable; either way the err != nil
		// branch of url.Parse is exercised.
		client := newClientForTest(t, "http://127.0.0.1:0")
		_, err := client.doRequest(ctx, http.MethodGet, "%zz", nil)
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// handleResponse error-path coverage
// ---------------------------------------------------------------------------

// errReader fails on Read so io.ReadAll surfaces an error.
type errReader struct{}

// reqproof:req REQ-API-020
func (errReader) Read(_ []byte) (int, error) { return 0, fmt.Errorf("boom") }

// reqproof:req REQ-API-020
func (errReader) Close() error { return nil }

// reqproof:req REQ-API-020
func TestClient_handleResponse_ErrorPaths(t *testing.T) {
	client := newClientForTest(t, "http://127.0.0.1:0")

	t.Run("body read failure", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       errReader{},
		}
		err := client.handleResponse(resp, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read response body")
	})

	t.Run("error response with structured json", func(t *testing.T) {
		// The Dashboard convention shape {"Status":"Error","Message":"..."}
		// fails to unmarshal into types.ErrorResponse (Status is a string,
		// not an int), so the fallback msgOnly path is exercised AND the
		// extracted message is non-empty.
		body := `{"Status":"Error","Message":"forbidden by policy","Meta":null}`
		resp := &http.Response{
			StatusCode: http.StatusForbidden,
			Status:     "403 Forbidden",
			Body:       io.NopCloser(strings.NewReader(body)),
		}
		err := client.handleResponse(resp, nil)
		require.Error(t, err)
		errResp, ok := err.(*types.ErrorResponse)
		require.True(t, ok)
		assert.Equal(t, 403, errResp.Status)
		assert.Contains(t, errResp.Message, "forbidden by policy")
		assert.Contains(t, errResp.Message, "403")
	})

	t.Run("error response with malformed body falls to status-text branch", func(t *testing.T) {
		// Not JSON at all, so both unmarshal attempts fail and the final
		// "status: body" formatting branch is exercised.
		resp := &http.Response{
			StatusCode: http.StatusBadGateway,
			Status:     "502 Bad Gateway",
			Body:       io.NopCloser(strings.NewReader("upstream went away")),
		}
		err := client.handleResponse(resp, nil)
		require.Error(t, err)
		errResp, ok := err.(*types.ErrorResponse)
		require.True(t, ok)
		assert.Equal(t, http.StatusBadGateway, errResp.Status)
		assert.Contains(t, errResp.Message, "502 Bad Gateway")
		assert.Contains(t, errResp.Message, "upstream went away")
	})

	t.Run("error response with json that has empty Message uses status fallback", func(t *testing.T) {
		// msgOnly unmarshal succeeds but Message field is empty, so the
		// `m.Message != ""` half of the short-circuit must be FALSE and
		// the function falls into the status-text branch.
		body := `{"Status":"Error","Meta":null}`
		resp := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Status:     "500 Internal Server Error",
			Body:       io.NopCloser(strings.NewReader(body)),
		}
		err := client.handleResponse(resp, nil)
		require.Error(t, err)
		errResp, ok := err.(*types.ErrorResponse)
		require.True(t, ok)
		assert.Equal(t, 500, errResp.Status)
		assert.Contains(t, errResp.Message, "500 Internal Server Error")
	})

	t.Run("success path with malformed json target body", func(t *testing.T) {
		// 200 OK but body is not valid JSON for the supplied target.
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("not-json")),
		}
		var out map[string]string
		err := client.handleResponse(resp, &out)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal response")
	})

	t.Run("success path with nil result skips unmarshal", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("anything goes")),
		}
		err := client.handleResponse(resp, nil)
		assert.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// GetOASAPI / CreateOASAPI / UpdateOASAPI / DeleteOASAPI error-path coverage
// ---------------------------------------------------------------------------

// reqproof:req REQ-API-002
func TestClient_GetOASAPI_VersionAndErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("version_name query parameter is appended", func(t *testing.T) {
		var gotQuery string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"info": map[string]interface{}{"title": "A"},
				"x-tyk-api-gateway": map[string]interface{}{
					"info": map[string]interface{}{"id": "x", "name": "X"},
				},
			})
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		api, err := client.GetOASAPI(ctx, "x", "v2")
		require.NoError(t, err)
		assert.Equal(t, "X", api.Name)
		assert.Contains(t, gotQuery, "version_name=v2")
	})

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.GetOASAPI(ctx, "x", "")
		require.Error(t, err)
	})

	t.Run("body read failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hj, ok := w.(http.Hijacker)
			require.True(t, ok)
			conn, _, err := hj.Hijack()
			require.NoError(t, err)
			// Send headers with Content-Length but close connection before body
			_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n"))
			_ = conn.Close()
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.GetOASAPI(ctx, "x", "")
		require.Error(t, err)
	})

	t.Run("error status with structured json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"status":404,"message":"api not found"}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.GetOASAPI(ctx, "x", "")
		require.Error(t, err)
		errResp, ok := err.(*types.ErrorResponse)
		require.True(t, ok)
		assert.Equal(t, 404, errResp.Status)
		assert.Contains(t, errResp.Message, "api not found")
	})

	t.Run("error status with malformed body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("oops"))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.GetOASAPI(ctx, "x", "")
		require.Error(t, err)
		errResp, ok := err.(*types.ErrorResponse)
		require.True(t, ok)
		assert.Equal(t, 500, errResp.Status)
	})

	t.Run("malformed oas document body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.GetOASAPI(ctx, "x", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal OAS document")
	})

	t.Run("parse metadata failure propagates", func(t *testing.T) {
		// Valid JSON but missing required "info" section so
		// parseOASDocumentToAPI returns an error.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"x-tyk-api-gateway":{"info":{}}}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.GetOASAPI(ctx, "x", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse API metadata")
	})
}

// reqproof:req REQ-API-003
func TestClient_CreateOASAPI_ErrorPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.CreateOASAPI(ctx, map[string]interface{}{"k": "v"})
		require.Error(t, err)
	})

	t.Run("handleResponse error from non-2xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":400,"message":"bad oas"}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.CreateOASAPI(ctx, map[string]interface{}{"k": "v"})
		require.Error(t, err)
	})

	t.Run("response missing API id", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// status 200 but ID field missing → result.ID == "" branch.
			_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success"})
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.CreateOASAPI(ctx, map[string]interface{}{"k": "v"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing API ID")
	})
}

// reqproof:req REQ-API-006
func TestClient_UpdateOASAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("success path follows up with GetOASAPI", func(t *testing.T) {
		var sawPut, sawGet bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPut:
				sawPut = true
				_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success", ID: "id-1"})
			case http.MethodGet:
				sawGet = true
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"info": map[string]interface{}{"title": "T"},
					"x-tyk-api-gateway": map[string]interface{}{
						"info": map[string]interface{}{"id": "id-1", "name": "T"},
					},
				})
			}
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		api, err := client.UpdateOASAPI(ctx, "id-1", map[string]interface{}{"k": "v"})
		require.NoError(t, err)
		assert.Equal(t, "id-1", api.ID)
		assert.True(t, sawPut)
		assert.True(t, sawGet)
	})

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.UpdateOASAPI(ctx, "x", map[string]interface{}{"k": "v"})
		require.Error(t, err)
	})

	t.Run("handleResponse error from 4xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":400,"message":"bad request"}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.UpdateOASAPI(ctx, "x", map[string]interface{}{"k": "v"})
		require.Error(t, err)
	})
}

// reqproof:req REQ-API-007
func TestClient_DeleteOASAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/api/apis/oas/abc", r.URL.Path)
			_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success", Message: "deleted"})
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		err := client.DeleteOASAPI(ctx, "abc")
		assert.NoError(t, err)
	})

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		err := client.DeleteOASAPI(ctx, "x")
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// ListOASAPIs / ListAPIsDashboard / ListOASAPIVersions / SwitchDefaultVersion
// ---------------------------------------------------------------------------

// reqproof:req REQ-API-001
func TestClient_ListOASAPIs_ZeroPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// page <= 0 must NOT append a `p=` query string.
		assert.Empty(t, r.URL.RawQuery)
		_ = json.NewEncoder(w).Encode(types.OASAPIListResponse{
			APIResponse: types.APIResponse{Status: "success"},
			APIs:        []*types.OASAPI{},
		})
	}))
	defer server.Close()

	client := newClientForTest(t, server.URL)
	apis, err := client.ListOASAPIs(context.Background(), 0)
	require.NoError(t, err)
	assert.Empty(t, apis)
}

// reqproof:req REQ-API-001
func TestClient_ListOASAPIs_ErrorPaths(t *testing.T) {
	ctx := context.Background()

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.ListOASAPIs(ctx, 0)
		require.Error(t, err)
	})

	t.Run("handleResponse error from 5xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status":500,"message":"server error"}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.ListOASAPIs(ctx, 0)
		require.Error(t, err)
	})
}

// reqproof:req REQ-API-001
func TestClient_ListAPIsDashboard(t *testing.T) {
	ctx := context.Background()

	t.Run("success with mixed api entries", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/apis", r.URL.Path)
			assert.Equal(t, "3", r.URL.Query().Get("p"))
			payload := map[string]interface{}{
				"apis": []interface{}{
					// Valid entry — should be included.
					map[string]interface{}{
						"api_definition": map[string]interface{}{
							"api_id": "a1",
							"name":   "Alpha",
							"proxy": map[string]interface{}{
								"listen_path": "/alpha",
							},
						},
					},
					// Not a map → outer ok=false branch.
					"not-a-map",
					// Map but no api_definition → second ok=false branch.
					map[string]interface{}{"other": 1},
					// api_definition is not a map → third ok=false branch.
					map[string]interface{}{"api_definition": "wrong-type"},
					// Valid api_definition but proxy is not a map → middle nested branch.
					map[string]interface{}{
						"api_definition": map[string]interface{}{
							"api_id": "a2",
							"name":   "Beta",
							"proxy":  "bad",
						},
					},
					// Valid api_definition with proxy map but listen_path missing.
					map[string]interface{}{
						"api_definition": map[string]interface{}{
							"api_id": "a3",
							"name":   "Gamma",
							"proxy":  map[string]interface{}{},
						},
					},
					// api_definition with empty api_id → final apiID != "" FALSE branch.
					map[string]interface{}{
						"api_definition": map[string]interface{}{
							"api_id": "",
							"name":   "Empty",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(payload)
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		apis, err := client.ListAPIsDashboard(ctx, 3)
		require.NoError(t, err)
		require.Len(t, apis, 3)
		assert.Equal(t, "a1", apis[0].ID)
		assert.Equal(t, "/alpha", apis[0].ListenPath)
		assert.Equal(t, "a2", apis[1].ID)
		assert.Empty(t, apis[1].ListenPath)
		assert.Equal(t, "a3", apis[2].ID)
	})

	t.Run("page <= 0 omits pagination query", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.URL.RawQuery)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"apis": []interface{}{}})
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		apis, err := client.ListAPIsDashboard(ctx, 0)
		require.NoError(t, err)
		assert.Empty(t, apis)
	})

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.ListAPIsDashboard(ctx, 0)
		require.Error(t, err)
	})

	t.Run("body read failure via early close", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hj, ok := w.(http.Hijacker)
			require.True(t, ok)
			conn, _, err := hj.Hijack()
			require.NoError(t, err)
			_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n"))
			_ = conn.Close()
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		client.SetTimeout(200 * time.Millisecond)
		_, err := client.ListAPIsDashboard(ctx, 0)
		require.Error(t, err)
	})

	t.Run("error status with json body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"status":401,"message":"nope"}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.ListAPIsDashboard(ctx, 0)
		require.Error(t, err)
		errResp, ok := err.(*types.ErrorResponse)
		require.True(t, ok)
		assert.Equal(t, 401, errResp.Status)
	})

	t.Run("error status with malformed body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("oops"))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.ListAPIsDashboard(ctx, 0)
		require.Error(t, err)
	})

	t.Run("malformed json body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not-json"))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.ListAPIsDashboard(ctx, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal dashboard")
	})

	t.Run("apis field missing or wrong type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"apis":"not-an-array"}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, err := client.ListAPIsDashboard(ctx, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "'apis' field not found")
	})
}

// reqproof:req REQ-API-002
func TestClient_ListOASAPIVersions(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/apis/oas/abc/versions", r.URL.Path)
			_ = json.NewEncoder(w).Encode(types.VersionListResponse{
				APIResponse: types.APIResponse{Status: "success"},
				Versions:    []string{"v1", "v2"},
				Default:     "v1",
			})
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		versions, def, err := client.ListOASAPIVersions(ctx, "abc")
		require.NoError(t, err)
		assert.Equal(t, []string{"v1", "v2"}, versions)
		assert.Equal(t, "v1", def)
	})

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		_, _, err := client.ListOASAPIVersions(ctx, "abc")
		require.Error(t, err)
	})

	t.Run("handleResponse error from 4xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"status":404,"message":"not found"}`))
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		_, _, err := client.ListOASAPIVersions(ctx, "abc")
		require.Error(t, err)
	})
}

// reqproof:req REQ-API-002
func TestClient_SwitchDefaultVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			body, _ := io.ReadAll(r.Body)
			var payload map[string]interface{}
			require.NoError(t, json.Unmarshal(body, &payload))
			assert.Equal(t, "v2", payload["set_default_version"])
			_ = json.NewEncoder(w).Encode(types.APIResponse{Status: "success"})
		}))
		defer server.Close()

		client := newClientForTest(t, server.URL)
		err := client.SwitchDefaultVersion(ctx, "abc", "v2")
		assert.NoError(t, err)
	})

	t.Run("doRequest network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
		closedURL := server.URL
		server.Close()

		client := newClientForTest(t, closedURL)
		client.SetTimeout(200 * time.Millisecond)
		err := client.SwitchDefaultVersion(ctx, "abc", "v2")
		require.Error(t, err)
	})
}

// reqproof:req REQ-CFG-002
func TestClient_Health_NetworkFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedURL := server.URL
	server.Close()

	client := newClientForTest(t, closedURL)
	client.SetTimeout(200 * time.Millisecond)
	err := client.Health(context.Background())
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// parseOASDocumentToAPI and getString coverage
// ---------------------------------------------------------------------------

// reqproof:req REQ-API-013
func TestClient_parseOASDocumentToAPI(t *testing.T) {
	client := newClientForTest(t, "http://127.0.0.1:0")

	t.Run("missing info section", func(t *testing.T) {
		_, err := client.parseOASDocumentToAPI(map[string]interface{}{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing info section")
	})

	t.Run("missing x-tyk-api-gateway extension", func(t *testing.T) {
		_, err := client.parseOASDocumentToAPI(map[string]interface{}{
			"info": map[string]interface{}{"title": "T"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "x-tyk-api-gateway")
	})

	t.Run("missing tyk info section", func(t *testing.T) {
		_, err := client.parseOASDocumentToAPI(map[string]interface{}{
			"info":              map[string]interface{}{"title": "T"},
			"x-tyk-api-gateway": map[string]interface{}{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing info in x-tyk-api-gateway")
	})

	t.Run("server present but listenPath wrong type", func(t *testing.T) {
		api, err := client.parseOASDocumentToAPI(map[string]interface{}{
			"info": map[string]interface{}{"title": "Fallback Title"},
			"x-tyk-api-gateway": map[string]interface{}{
				"info":   map[string]interface{}{"id": "x"},
				"server": map[string]interface{}{"listenPath": "bad-type"},
			},
		})
		require.NoError(t, err)
		assert.Empty(t, api.ListenPath)
		// Name was empty in tyk info so falls back to OAS info title.
		assert.Equal(t, "Fallback Title", api.Name)
	})

	t.Run("listenPath present but value wrong type", func(t *testing.T) {
		api, err := client.parseOASDocumentToAPI(map[string]interface{}{
			"info": map[string]interface{}{"title": "T"},
			"x-tyk-api-gateway": map[string]interface{}{
				"info": map[string]interface{}{"id": "x", "name": "Named"},
				"server": map[string]interface{}{
					"listenPath": map[string]interface{}{"value": 42},
				},
			},
		})
		require.NoError(t, err)
		assert.Empty(t, api.ListenPath)
		// Name from tyk info is non-empty so the title fallback is NOT taken
		// — that exercises the api.Name == "" FALSE branch.
		assert.Equal(t, "Named", api.Name)
	})

	t.Run("upstream wrong type", func(t *testing.T) {
		api, err := client.parseOASDocumentToAPI(map[string]interface{}{
			"info": map[string]interface{}{"title": "T"},
			"x-tyk-api-gateway": map[string]interface{}{
				"info":     map[string]interface{}{"id": "x", "name": "N"},
				"upstream": "bad",
			},
		})
		require.NoError(t, err)
		assert.Empty(t, api.UpstreamURL)
	})

	t.Run("upstream present but url wrong type", func(t *testing.T) {
		api, err := client.parseOASDocumentToAPI(map[string]interface{}{
			"info": map[string]interface{}{"title": "T"},
			"x-tyk-api-gateway": map[string]interface{}{
				"info":     map[string]interface{}{"id": "x", "name": "N"},
				"upstream": map[string]interface{}{"url": 7},
			},
		})
		require.NoError(t, err)
		assert.Empty(t, api.UpstreamURL)
	})
}

// reqproof:req REQ-API-013
func TestClient_getString(t *testing.T) {
	t.Run("returns value when string", func(t *testing.T) {
		assert.Equal(t, "v", getString(map[string]interface{}{"k": "v"}, "k"))
	})
	t.Run("returns empty when missing", func(t *testing.T) {
		assert.Empty(t, getString(map[string]interface{}{}, "k"))
	})
	t.Run("returns empty when wrong type", func(t *testing.T) {
		assert.Empty(t, getString(map[string]interface{}{"k": 42}, "k"))
	})
}

// reqproof:req REQ-API-030
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
