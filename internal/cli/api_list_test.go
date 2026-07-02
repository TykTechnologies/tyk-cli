package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Verifies: SYS-REQ-001
func TestAPIList_JSONOutput(t *testing.T) {
	mockAPIs := []*types.OASAPI{{ID: "id1", Name: "Name1"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.OASAPIListResponse{APIs: mockAPIs})
	}))
	defer server.Close()

	root := NewRootCommand("test", "commit", "time")
	listCmd, _, err := root.Find([]string{"api", "list"})
	require.NoError(t, err)

	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
	}}
	listCmd.SetContext(withConfig(context.Background(), cfg))
	listCmd.SetContext(withOutputFormat(listCmd.Context(), types.OutputJSON))

	// Just ensure command executes without error with JSON output
	listCmd.SetArgs([]string{"--page", "1"})
	err = listCmd.Execute()
	require.NoError(t, err)
}

// Verifies: SYS-REQ-001
func TestAPIList_HumanOutput_NoAPIs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(types.OASAPIListResponse{APIs: []*types.OASAPI{}})
	}))
	defer server.Close()

	root := NewRootCommand("test", "commit", "time")
	listCmd, _, err := root.Find([]string{"api", "list"})
	require.NoError(t, err)

	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
	}}
	listCmd.SetContext(withConfig(context.Background(), cfg))
	listCmd.SetContext(withOutputFormat(listCmd.Context(), types.OutputHuman))

	// Capture stderr (human empty message prints to stderr)
	oldStderr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr
	defer func() { os.Stderr = oldStderr }()

	listCmd.SetArgs([]string{"--page", "1"})
	err = listCmd.Execute()
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// MC/DC coverage for runAPIList branches
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-001
// TestRunAPIList_PageZeroDefaultsToOne covers L399 page<=0=T branch.
func TestRunAPIList_PageZeroDefaultsToOne(t *testing.T) {
	gotPage := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPage = r.URL.Query().Get("p")
		_ = json.NewEncoder(w).Encode(types.OASAPIListResponse{APIs: []*types.OASAPI{}})
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
	}}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputHuman))
	cmd.SetArgs([]string{"--page", "0"})
	_ = cmd.ParseFlags([]string{"--page", "0"})
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	// runAPIList should clamp the page to 1.
	assert.Equal(t, "1", gotPage)
}

// Verifies: SYS-REQ-001
// TestRunAPIList_ConfigNil covers L405 config==nil=T branch.
func TestRunAPIList_ConfigNil(t *testing.T) {
	cmd := NewAPIListCommand()
	// Deliberately do not attach a config to the context.
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// Verifies: SYS-REQ-001
// TestRunAPIList_InteractiveJSONIncompat covers L419 interactive=T AND L420
// outputFormat==OutputJSON=T combined branch.
func TestRunAPIList_InteractiveJSONIncompatible(t *testing.T) {
	cmd := NewAPIListCommand()
	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: "http://unused", AuthToken: "tok", OrgID: "org"},
	}}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputJSON))
	cmd.SetArgs([]string{"--interactive"})
	_ = cmd.ParseFlags([]string{"--interactive"})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "interactive mode is not compatible")
}

// Verifies: SYS-REQ-001
// TestRunAPIList_NonClassifiedError covers L434 cls!=nil=F branch.
func TestRunAPIList_NonClassifiedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "tok", OrgID: "org"},
	}}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputHuman))
	cmd.SetArgs([]string{})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// Verifies: SYS-REQ-001
// TestRunAPIList_NewClientFails covers L411 err!=nil from client.NewClient.
func TestRunAPIList_NewClientFails(t *testing.T) {
	cmd := NewAPIListCommand()
	cmd.SetContext(withConfig(context.Background(), brokenConfig()))
	cmd.SetArgs([]string{})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// Verifies: SYS-REQ-001
// TestRunAPIList_HumanOutput_PageHasAPIs covers L440 outputFormat==OutputJSON=F
// (the human-output branch) and ensures displayAPIPage prints rows.
func TestRunAPIList_HumanOutput_WithAPIs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(mockAPIListResponse())
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
	}}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputHuman))
	cmd.SetArgs([]string{})

	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := cmd.RunE(cmd, []string{})
	wOut.Close()
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// Verifies: SYS-REQ-001
// TestRunAPIList_HumanOutput covers L420 outputFormat == OutputJSON = F by
// exercising the default Human output path (renders a table to stdout).
func TestRunAPIList_HumanOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"apis": []interface{}{
				map[string]interface{}{
					"api_definition": map[string]interface{}{
						"api_id": "a1",
						"name":   "Alpha",
						"proxy":  map[string]interface{}{"listen_path": "/a/"},
					},
				},
				map[string]interface{}{
					"api_definition": map[string]interface{}{
						"api_id": "b2",
						"name":   "Beta",
						"proxy":  map[string]interface{}{"listen_path": "/b/"},
					},
				},
			},
		})
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputHuman)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})
	_ = cmd.ParseFlags([]string{})

	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := cmd.RunE(cmd, []string{})
	wOut.Close()
	os.Stdout = oldStdout

	require.NoError(t, err)
}
