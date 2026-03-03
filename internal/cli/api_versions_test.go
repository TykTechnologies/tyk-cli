package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func createVersionConfig(serverURL string) *types.Config {
	return &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {
				Name:         "test",
				DashboardURL: serverURL,
				AuthToken:    "test-token",
				OrgID:        "org",
			},
		},
	}
}

func executeVersionsListCmd(t *testing.T, serverURL string, outputFormat types.OutputFormat, apiID string) error {
	t.Helper()
	root := NewRootCommand("test", "commit", "time")
	listCmd, _, err := root.Find([]string{"api", "versions", "list"})
	require.NoError(t, err)

	cfg := createVersionConfig(serverURL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, outputFormat)
	listCmd.SetContext(ctx)
	listCmd.SetArgs([]string{"--api-id", apiID})
	_ = listCmd.ParseFlags([]string{"--api-id", apiID})

	return listCmd.RunE(listCmd, []string{})
}

// ---------------------------------------------------------------------------
// Test scenarios
// ---------------------------------------------------------------------------

// Test Budget: 3 behaviors x 2 = 6 max unit tests. Using 3.

func TestAPIVersionsList_Human(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/abc123/versions" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"versions": []string{"v1", "v2", "v3"},
				"default":  "v1",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Capture stderr (human output goes to stderr)
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := executeVersionsListCmd(t, server.URL, types.OutputHuman, "abc123")

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)

	require.NoError(t, err)
	output := string(stderr)
	assert.Contains(t, output, "v1")
	assert.Contains(t, output, "v2")
	assert.Contains(t, output, "v3")
	assert.Contains(t, output, "* v1", "default version should have marker")
	assert.Contains(t, output, "3 version(s) found")
	assert.Contains(t, output, "Default: v1")
}

func TestAPIVersionsList_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/xyz789/versions" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"versions": []string{"v1", "v2"},
				"default":  "v2",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Capture stdout (JSON goes to stdout)
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executeVersionsListCmd(t, server.URL, types.OutputJSON, "xyz789")

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(stdout, &result)
	require.NoError(t, err, "output should be valid JSON")
	assert.Equal(t, "xyz789", result["api_id"])
	assert.Equal(t, "v2", result["default"])
	versions, ok := result["versions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, versions, 2)
}

func TestAPIVersionsList_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  404,
			"message": "API not found",
		})
	}))
	defer server.Close()

	err := executeVersionsListCmd(t, server.URL, types.OutputHuman, "nonexistent")

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, exitErr.Message, "not found")
}

// ===========================================================================
// Switch Default tests
// ===========================================================================

// Test Budget: 5 behaviors x 2 = 10 max unit tests. Using 6.
// Behaviors: success switch, no-op already default, version not found, JSON output, JSON no-op, API not found

func executeVersionsSwitchDefaultCmd(t *testing.T, serverURL string, outputFormat types.OutputFormat, apiID, versionName string) error {
	t.Helper()
	cmd := NewAPIVersionsSwitchDefaultCommand()

	cfg := createVersionConfig(serverURL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, outputFormat)
	cmd.SetContext(ctx)

	args := []string{"--api-id", apiID, "--version-name", versionName}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)

	return cmd.RunE(cmd, []string{})
}

// mockSwitchServer creates a test server for switch-default tests.
func mockSwitchServer(t *testing.T, versions []string, defaultVersion string, patchCalled *bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis/oas/api-123/versions":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"versions": versions,
				"default":  defaultVersion,
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/apis/oas/api-123":
			if patchCalled != nil {
				*patchCalled = true
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "OK"})
		default:
			t.Logf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
}

func TestAPIVersionsSwitchDefault_Success(t *testing.T) {
	patchCalled := false
	server := mockSwitchServer(t, []string{"v1", "v2", "v3"}, "v1", &patchCalled)
	defer server.Close()

	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := executeVersionsSwitchDefaultCmd(t, server.URL, types.OutputHuman, "api-123", "v3")

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)

	require.NoError(t, err)
	assert.True(t, patchCalled, "PATCH should have been called to switch default")
	stderrStr := string(stderr)
	assert.Contains(t, stderrStr, "v1", "stderr should show previous default")
	assert.Contains(t, stderrStr, "v3", "stderr should show new default")
}

func TestAPIVersionsSwitchDefault_AlreadyDefault(t *testing.T) {
	patchCalled := false
	server := mockSwitchServer(t, []string{"v1", "v2"}, "v1", &patchCalled)
	defer server.Close()

	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := executeVersionsSwitchDefaultCmd(t, server.URL, types.OutputHuman, "api-123", "v1")

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)

	require.NoError(t, err, "no-op should return nil (exit 0)")
	assert.False(t, patchCalled, "PATCH should NOT be called when already default")
	assert.Contains(t, string(stderr), "already the default")
}

func TestAPIVersionsSwitchDefault_VersionNotFound(t *testing.T) {
	server := mockSwitchServer(t, []string{"v1", "v2", "v3"}, "v1", nil)
	defer server.Close()

	err := executeVersionsSwitchDefaultCmd(t, server.URL, types.OutputHuman, "api-123", "v99")

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, exitErr.Message, "v99")
	assert.Contains(t, exitErr.Message, "v1")
}

func TestAPIVersionsSwitchDefault_JSONOutput(t *testing.T) {
	patchCalled := false
	server := mockSwitchServer(t, []string{"v1", "v2", "v3"}, "v1", &patchCalled)
	defer server.Close()

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executeVersionsSwitchDefaultCmd(t, server.URL, types.OutputJSON, "api-123", "v3")

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)
	assert.True(t, patchCalled)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(stdout, &result), "output should be valid JSON")
	assert.Equal(t, "switched", result["action"])
	assert.Equal(t, "v1", result["previous_default"])
	assert.Equal(t, "v3", result["new_default"])
	assert.Equal(t, "api-123", result["api_id"])
}

func TestAPIVersionsSwitchDefault_AlreadyDefault_JSON(t *testing.T) {
	patchCalled := false
	server := mockSwitchServer(t, []string{"v1", "v2"}, "v1", &patchCalled)
	defer server.Close()

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executeVersionsSwitchDefaultCmd(t, server.URL, types.OutputJSON, "api-123", "v1")

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)
	assert.False(t, patchCalled)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(stdout, &result))
	assert.Equal(t, "no_change", result["action"])
	assert.Equal(t, "v1", result["previous_default"])
	assert.Equal(t, "v1", result["new_default"])
}

func TestAPIVersionsSwitchDefault_APINotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  404,
			"message": fmt.Sprintf("API not found: %s", "api-missing"),
		})
	}))
	defer server.Close()

	err := executeVersionsSwitchDefaultCmd(t, server.URL, types.OutputHuman, "api-missing", "v1")

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, exitErr.Message, "api-missing")
}
