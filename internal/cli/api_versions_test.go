package cli

import (
	"context"
	"encoding/json"
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
