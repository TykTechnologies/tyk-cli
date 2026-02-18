package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Mock OAS API response with Tyk extensions
func mockOASAPIResponse() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":   "Test API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{
			"/test": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Test endpoint",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Success",
						},
					},
				},
			},
		},
		"x-tyk-api-gateway": map[string]interface{}{
			"info": map[string]interface{}{
				"id":   "test-api-id",
				"name": "Test API",
			},
			"server": map[string]interface{}{
				"listenPath": map[string]interface{}{
					"value": "/test-api/",
				},
			},
			"upstream": map[string]interface{}{
				"url": "http://upstream.example.com",
			},
		},
	}
}

func TestAPIGet_WithOASOnly_JSON(t *testing.T) {
	mockOAS := mockOASAPIResponse()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/apis/oas/test-api-id", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOAS)
	}))
	defer server.Close()

	// Create the command directly
	getCmd := NewAPIGetCommand()
	
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	getCmd.SetContext(withConfig(context.Background(), cfg))
	getCmd.SetContext(withOutputFormat(getCmd.Context(), types.OutputJSON))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command with --oas-only flag
	getCmd.SetArgs([]string{"test-api-id", "--oas-only"})
	err := getCmd.Execute()
	require.NoError(t, err)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)

	// Parse JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	// Verify that x-tyk-api-gateway is NOT present
	_, hasTykExt := result["x-tyk-api-gateway"]
	assert.False(t, hasTykExt, "x-tyk-api-gateway should not be present in OAS-only output")

	// Verify that standard OAS fields are present
	assert.Equal(t, "3.0.3", result["openapi"])
	assert.NotNil(t, result["info"])
	assert.NotNil(t, result["paths"])
}

func TestAPIGet_WithOASOnly_YAML(t *testing.T) {
	mockOAS := mockOASAPIResponse()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/apis/oas/test-api-id", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOAS)
	}))
	defer server.Close()

	// Create the command directly
	getCmd := NewAPIGetCommand()

	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	getCmd.SetContext(withConfig(context.Background(), cfg))
	getCmd.SetContext(withOutputFormat(getCmd.Context(), types.OutputHuman))

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Execute command with --oas-only flag
	getCmd.SetArgs([]string{"test-api-id", "--oas-only"})
	err := getCmd.Execute()
	require.NoError(t, err)

	// Restore stdout/stderr and read output
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout, _ := io.ReadAll(rOut)
	stderr, _ := io.ReadAll(rErr)

	yamlOutput := string(stdout)
	stderrOutput := string(stderr)

	// Verify that x-tyk-api-gateway is NOT present in YAML output
	assert.NotContains(t, yamlOutput, "x-tyk-api-gateway", "x-tyk-api-gateway should not be present in OAS-only YAML output")

	// Verify that standard OAS fields are present
	assert.Contains(t, yamlOutput, "openapi: 3.0.3")
	assert.Contains(t, yamlOutput, "title: Test API")
	assert.Contains(t, yamlOutput, "paths:")

	// Verify that API summary is NOT present (should be empty stderr in OAS-only mode)
	assert.Empty(t, stderrOutput, "No API summary should be shown in OAS-only mode")
}

func TestAPIGet_WithoutOASOnly_ShowsFullOutput(t *testing.T) {
	mockOAS := mockOASAPIResponse()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/apis/oas/test-api-id", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOAS)
	}))
	defer server.Close()

	// Create the command directly
	getCmd := NewAPIGetCommand()

	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	getCmd.SetContext(withConfig(context.Background(), cfg))
	getCmd.SetContext(withOutputFormat(getCmd.Context(), types.OutputJSON))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute command WITHOUT --oas-only flag
	getCmd.SetArgs([]string{"test-api-id"})
	err := getCmd.Execute()
	require.NoError(t, err)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)

	// Parse JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)

	// Verify this is the full API response structure
	assert.NotNil(t, result["id"])
	assert.NotNil(t, result["name"])
	assert.NotNil(t, result["oas"])

	// Verify that OAS contains x-tyk-api-gateway
	oasData, ok := result["oas"].(map[string]interface{})
	require.True(t, ok, "OAS field should be a map")
	_, hasTykExt := oasData["x-tyk-api-gateway"]
	assert.True(t, hasTykExt, "x-tyk-api-gateway should be present in normal output")
}

func TestAPIGet_WithOASOnly_HumanOutput_ShowsNoSummary(t *testing.T) {
	mockOAS := mockOASAPIResponse()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOAS)
	}))
	defer server.Close()

	// Create the command directly
	getCmd := NewAPIGetCommand()

	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	getCmd.SetContext(withConfig(context.Background(), cfg))
	getCmd.SetContext(withOutputFormat(getCmd.Context(), types.OutputHuman))

	// Capture stderr where API summary would be written
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Capture stdout where YAML would be written
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	getCmd.SetArgs([]string{"test-api-id", "--oas-only"})
	err := getCmd.Execute()
	require.NoError(t, err)

	// Restore and read outputs
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout

	stderr, _ := io.ReadAll(rErr)
	stdout, _ := io.ReadAll(rOut)

	// Verify no API summary is shown in stderr
	stderrStr := strings.TrimSpace(string(stderr))
	assert.Empty(t, stderrStr, "No API summary should be shown in --oas-only mode")

	// Verify clean YAML output in stdout
	stdoutStr := string(stdout)
	assert.Contains(t, stdoutStr, "openapi: 3.0.3")
	assert.NotContains(t, stdoutStr, "x-tyk-api-gateway")
}

func TestAPIGet_WithoutOASOnly_HumanOutput_ShowsSummary(t *testing.T) {
	mockOAS := mockOASAPIResponse()
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockOAS)
	}))
	defer server.Close()

	// Create the command directly
	getCmd := NewAPIGetCommand()

	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	getCmd.SetContext(withConfig(context.Background(), cfg))
	getCmd.SetContext(withOutputFormat(getCmd.Context(), types.OutputHuman))

	// Capture stderr where API summary is written
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Capture stdout for YAML
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	getCmd.SetArgs([]string{"test-api-id"})
	err := getCmd.Execute()
	require.NoError(t, err)

	// Restore outputs
	wOut.Close()
	wErr.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout

	stdout, _ := io.ReadAll(rOut)
	stderr, _ := io.ReadAll(rErr)
	stderrStr := string(stderr)

	// Verify API summary is shown in stderr
	assert.Contains(t, stderrStr, "API Summary:")
	assert.Contains(t, stderrStr, "ID:")
	assert.Contains(t, stderrStr, "Name:")
	assert.Contains(t, stderrStr, "OpenAPI Specification:")

	// Verify full YAML output in stdout (including x-tyk-api-gateway)
	stdoutStr := string(stdout)
	assert.Contains(t, stdoutStr, "openapi: 3.0.3")
	assert.Contains(t, stdoutStr, "x-tyk-api-gateway")
}

func TestAPIGet_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("API not found"))
	}))
	defer server.Close()

	// Create the command directly
	getCmd := NewAPIGetCommand()

	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	getCmd.SetContext(withConfig(context.Background(), cfg))
	getCmd.SetContext(withOutputFormat(getCmd.Context(), types.OutputJSON))

	// This should fail with a 404 error
	getCmd.SetArgs([]string{"non-existent-api", "--oas-only"})
	err := getCmd.Execute()
	require.Error(t, err)

	// Check that it's the expected error type
	if exitErr, ok := err.(*ExitError); ok {
		assert.Equal(t, 3, exitErr.Code)
		assert.Contains(t, exitErr.Message, "not found")
	} else {
		assert.Contains(t, err.Error(), "not found")
	}
}