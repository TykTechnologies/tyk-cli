package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
	"gopkg.in/yaml.v3"
)

// mockCleanOAS creates a clean OpenAPI spec without Tyk extensions
func mockCleanOAS() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Clean Test API",
			"version":     "1.0.0",
			"description": "A clean OpenAPI spec for testing",
		},
		"paths": map[string]interface{}{
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get users",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Success",
						},
					},
				},
			},
		},
		"servers": []interface{}{
			map[string]interface{}{
				"url": "https://api.example.com",
			},
		},
	}
}

// mockTykEnhancedOAS creates a Tyk-enhanced OAS with extensions
func mockTykEnhancedOAS() map[string]interface{} {
	cleanOAS := mockCleanOAS()
	cleanOAS["x-tyk-api-gateway"] = map[string]interface{}{
		"info": map[string]interface{}{
			"id":   "test-api-123",
			"name": "Enhanced Test API",
		},
		"server": map[string]interface{}{
			"listenPath": map[string]interface{}{
				"value": "/enhanced-api/",
			},
		},
		"upstream": map[string]interface{}{
			"url": "https://api.example.com",
		},
	}
	return cleanOAS
}

// mockCreateAPIResponse simulates the API creation response
func mockCreateAPIResponse() types.APIResponse {
	return types.APIResponse{
		ID:      "new-api-456",
		Message: "API created successfully",
	}
}

// mockCreatedOASAPI simulates a created API with OAS data
func mockCreatedOASAPI() *types.OASAPI {
	return &types.OASAPI{
		ID:             "new-api-456",
		Name:           "Clean Test API",
		ListenPath:     "/clean-test-api/",
		DefaultVersion: "v1",
		UpstreamURL:    "https://api.example.com",
		OAS:            mockTykEnhancedOAS(), // API gets enhanced with Tyk extensions
	}
}

// createTempOASFile creates a temporary OAS file for testing
func createTempOASFile(t *testing.T, oasData map[string]interface{}) string {
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-api.yaml")

	// Convert to YAML and write
	yamlData, err := yaml.Marshal(oasData)
	require.NoError(t, err)

	err = os.WriteFile(tmpFile, yamlData, 0644)
	require.NoError(t, err)

	return tmpFile
}

func TestNewAPIImportOASCommand(t *testing.T) {
	cmd := NewAPIImportOASCommand()

	assert.Equal(t, "import-oas", cmd.Use)
	assert.Equal(t, "Import clean OpenAPI spec to create new API", cmd.Short)
	assert.Contains(t, cmd.Long, "Import a clean OpenAPI specification")
	assert.Contains(t, cmd.Long, "automatically generated Tyk extensions")
	
	// Check flags exist
	assert.True(t, cmd.Flags().Lookup("file") != nil)
	assert.True(t, cmd.Flags().Lookup("url") != nil)
}

func TestRunAPIImportOAS_WithFile(t *testing.T) {
	// Create a mock server that simulates API creation
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/api/apis/oas") {
			// Simulate API creation
			createResp := mockCreateAPIResponse()
			_ = json.NewEncoder(w).Encode(createResp)
		} else if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/api/apis/oas/new-api-456") {
			// Simulate getting the created API details
			api := mockCreatedOASAPI()
			// Return the OAS document directly as the API endpoint does
			_ = json.NewEncoder(w).Encode(api.OAS)
		}
	}))
	defer server.Close()

	// Create a temp OAS file
	cleanOAS := mockCleanOAS()
	tmpFile := createTempOASFile(t, cleanOAS)

	// Create command with context
	cmd := NewAPIImportOASCommand()
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), config))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputJSON))

	// Set flags
	_ = cmd.Flags().Set("file", tmpFile)

	// Execute command
	err := cmd.Execute()

	// Verify no error
	assert.NoError(t, err)
}

func TestRunAPIImportOAS_MissingInput(t *testing.T) {
	cmd := NewAPIImportOASCommand()
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: "http://test", AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), config))

	// Don't set file or url flags
	err := cmd.Execute()

	// Should get error about missing input
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Either --file or --url must be provided")
}

func TestRunAPIImportOAS_BothInputs(t *testing.T) {
	cmd := NewAPIImportOASCommand()
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: "http://test", AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), config))

	// Set both file and url flags
	_ = cmd.Flags().Set("file", "/tmp/test.yaml")
	_ = cmd.Flags().Set("url", "https://example.com/api.yaml")

	err := cmd.Execute()

	// Should get error about conflicting inputs
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot specify both --file and --url")
}

func TestNewAPIUpdateOASCommand(t *testing.T) {
	cmd := NewAPIUpdateOASCommand()

	assert.Equal(t, "update-oas <api-id>", cmd.Use)
	assert.Equal(t, "Update existing API's OpenAPI spec only", cmd.Short)
	assert.Contains(t, cmd.Long, "Update an existing API's OpenAPI specification")
	assert.Contains(t, cmd.Long, "preserving Tyk configuration")
	
	// Check flags exist
	assert.True(t, cmd.Flags().Lookup("file") != nil)
	assert.True(t, cmd.Flags().Lookup("url") != nil)
}

func TestRunAPIUpdateOAS_Success(t *testing.T) {
	testAPIID := "existing-api-123"
	
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, testAPIID) {
			// Simulate getting existing API
			existingAPI := mockCreatedOASAPI()
			existingAPI.ID = testAPIID
			_ = json.NewEncoder(w).Encode(existingAPI.OAS)
		} else if r.Method == http.MethodPut && strings.Contains(r.URL.Path, testAPIID) {
			// Simulate API update
			updateResp := types.APIResponse{ID: testAPIID, Message: "Updated"}
			_ = json.NewEncoder(w).Encode(updateResp)
		}
	}))
	defer server.Close()

	// Create a temp OAS file
	cleanOAS := mockCleanOAS()
	tmpFile := createTempOASFile(t, cleanOAS)

	// Create command with context
	cmd := NewAPIUpdateOASCommand()
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), config))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputJSON))

	// Set args and flags
	cmd.SetArgs([]string{testAPIID})
	_ = cmd.Flags().Set("file", tmpFile)

	// Execute command
	err := cmd.Execute()

	// Verify no error
	assert.NoError(t, err)
}

func TestRunAPIUpdateOAS_MissingAPIID(t *testing.T) {
	cmd := NewAPIUpdateOASCommand()
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: "http://test", AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), config))

	// Don't provide API ID argument
	err := cmd.Execute()

	// Should get error about missing argument
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestNewAPIApplyCommand_Enhanced(t *testing.T) {
	cmd := NewAPIApplyCommand()

	assert.Equal(t, "apply", cmd.Use)
	assert.Equal(t, "Apply Tyk-enhanced API configuration", cmd.Short)
	assert.Contains(t, cmd.Long, "Tyk-enhanced API configuration")
	assert.Contains(t, cmd.Long, "x-tyk-api-gateway extensions")
	assert.Contains(t, cmd.Long, "infrastructure-as-code")
	
    // Check enhanced help text
    assert.Contains(t, cmd.Long, "tyk api import-oas")
    assert.Contains(t, cmd.Long, "tyk api update-oas")
}

func TestRunAPIApply_PlainOASRejection(t *testing.T) {
	// Create a temp file with clean (non-Tyk-enhanced) OAS
	cleanOAS := mockCleanOAS()
	tmpFile := createTempOASFile(t, cleanOAS)

	// Create command
	cmd := NewAPIApplyCommand()
	config := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: "http://test", AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), config))

	// Set file flag
	_ = cmd.Flags().Set("file", tmpFile)

	// Execute command
	err := cmd.Execute()

	// Should get enhanced error message about missing extensions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lacks required x-tyk-api-gateway extensions")
	assert.Contains(t, err.Error(), "tyk api import-oas")
	assert.Contains(t, err.Error(), "tyk api update-oas")
}

func TestRunAPIApply_MissingIDCreatesAPI(t *testing.T) {
    // Create Tyk-enhanced OAS but without API ID
    enhancedOAS := mockTykEnhancedOAS()
    if tykExt, ok := enhancedOAS["x-tyk-api-gateway"].(map[string]interface{}); ok {
        if info, ok := tykExt["info"].(map[string]interface{}); ok {
            delete(info, "id")
        }
    }

    tmpFile := createTempOASFile(t, enhancedOAS)

    // Mock server for creation + fetch
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/api/apis/oas") {
            createResp := mockCreateAPIResponse()
            _ = json.NewEncoder(w).Encode(createResp)
            return
        }
        if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/api/apis/oas/new-api-456") {
            api := mockCreatedOASAPI()
            _ = json.NewEncoder(w).Encode(api.OAS)
            return
        }
        http.NotFound(w, r)
    }))
    defer server.Close()

    cmd := NewAPIApplyCommand()
    config := &types.Config{
        DefaultEnvironment: "test",
        Environments: map[string]*types.Environment{
            "test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
        },
    }
    cmd.SetContext(withConfig(context.Background(), config))

    _ = cmd.Flags().Set("file", tmpFile)

    // Execute command: should succeed and create new API
    err := cmd.Execute()
    assert.NoError(t, err)
}

func TestLoadOASFromFile_Success(t *testing.T) {
	// Create test OAS data
	testOAS := mockCleanOAS()
	tmpFile := createTempOASFile(t, testOAS)

	// Test the helper function
	loadedOAS, err := loadOASFromFile(tmpFile)

	// Verify success
	require.NoError(t, err)
	assert.Equal(t, "Clean Test API", loadedOAS["info"].(map[string]interface{})["title"])
	assert.Equal(t, "3.0.3", loadedOAS["openapi"])
}

func TestLoadOASFromFile_NotFound(t *testing.T) {
	// Test with non-existent file
	_, err := loadOASFromFile("/nonexistent/path/api.yaml")

	// Should get file not found error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestLoadOASFromURL_Success(t *testing.T) {
	// Create a test server that serves OAS
	testOAS := mockCleanOAS()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testOAS)
	}))
	defer server.Close()

	// Test the helper function
	loadedOAS, err := loadOASFromURL(server.URL + "/api.json")

	// Verify success
	require.NoError(t, err)
	assert.Equal(t, "Clean Test API", loadedOAS["info"].(map[string]interface{})["title"])
	assert.Equal(t, "3.0.3", loadedOAS["openapi"])
}

func TestLoadOASFromURL_HTTPError(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Test the helper function
	_, err := loadOASFromURL(server.URL + "/nonexistent.json")

	// Should get HTTP error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestLoadOASFromURL_InvalidJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("invalid json content"))
	}))
	defer server.Close()

	// Test the helper function
	_, err := loadOASFromURL(server.URL + "/invalid.json")

	// Should get parse error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse OAS document")
}
