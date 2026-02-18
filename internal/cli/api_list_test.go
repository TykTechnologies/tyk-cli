package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

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
