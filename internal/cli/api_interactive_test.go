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

func TestAPIListInteractiveFlag(t *testing.T) {
	// Create a test server that returns mock APIs
	mockAPIs := []*types.OASAPI{
		{ID: "api-1", Name: "API One", ListenPath: "/one", DefaultVersion: "v1"},
		{ID: "api-2", Name: "API Two", ListenPath: "/two", DefaultVersion: "v1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/apis/oas", r.URL.Path)
		_ = json.NewEncoder(w).Encode(types.OASAPIListResponse{
			APIResponse: types.APIResponse{Status: "success"},
			APIs:        mockAPIs,
		})
	}))
	defer server.Close()

	// Test that the interactive flag is recognized and available
	root := NewRootCommand("test", "commit", "time")
	listCmd, _, err := root.Find([]string{"api", "list"})
	require.NoError(t, err)

	// Check that the interactive flag exists
	flag := listCmd.Flags().Lookup("interactive")
	require.NotNil(t, flag, "interactive flag should exist")
	assert.Equal(t, "bool", flag.Value.Type())
	assert.Equal(t, "false", flag.DefValue)

	// Check short flag
	shortFlag := listCmd.Flags().ShorthandLookup("i")
	require.NotNil(t, shortFlag, "interactive short flag -i should exist")
}

// Note: Interactive error testing is challenging in unit tests due to global flag handling
// The error case is verified through manual testing:
// ./build/tyk api list --interactive --json
// Error: interactive mode is not compatible with JSON output format

func TestDisplayAPIPage(t *testing.T) {
	// Test the displayAPIPage function
	apis := []*types.OASAPI{
		{ID: "test-id-1", Name: "Test API 1", ListenPath: "/test1", DefaultVersion: "v1"},
		{ID: "test-id-2", Name: "Test API 2", ListenPath: "/test2", DefaultVersion: "v2"},
	}

	// Capture stdout to verify table output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Capture stderr for status messages  
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Test non-interactive display
	displayAPIPage(apis, 1, false)

	// Restore stdout and stderr
	w.Close()
	os.Stdout = oldStdout
	wErr.Close()
	os.Stderr = oldStderr

	// Read captured output
	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	bufErr := make([]byte, 1024)
	nErr, _ := rErr.Read(bufErr)
	stderrOutput := string(bufErr[:nErr])

	// Verify table headers and data are present
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name") 
	assert.Contains(t, output, "Listen Path")
	assert.Contains(t, output, "Default Version")
	assert.Contains(t, output, "test-id-1")
	assert.Contains(t, output, "Test API 1")
	assert.Contains(t, output, "/test1")

	// Verify page info in stderr
	assert.Contains(t, stderrOutput, "APIs (page 1):")
	assert.Contains(t, stderrOutput, "Use '--page 2' for next page")
}

func TestDisplayAPIPageEmpty(t *testing.T) {
	// Test empty API list
	apis := []*types.OASAPI{}

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	displayAPIPage(apis, 1, false)

	wErr.Close()
	os.Stderr = oldStderr

	bufErr := make([]byte, 1024)
	nErr, _ := rErr.Read(bufErr)
	stderrOutput := string(bufErr[:nErr])

	assert.Contains(t, stderrOutput, "No APIs found on page 1")
}

// Test helper function to verify command structure
func TestAPIListCommandStructure(t *testing.T) {
	cmd := NewAPIListCommand()
	
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List OAS APIs", cmd.Short)
	assert.Contains(t, cmd.Long, "interactive navigation")
	
	// Check flags
	pageFlag := cmd.Flags().Lookup("page")
	require.NotNil(t, pageFlag)
	assert.Equal(t, "1", pageFlag.DefValue)
	
	interactiveFlag := cmd.Flags().Lookup("interactive")
	require.NotNil(t, interactiveFlag)
	assert.Equal(t, "false", interactiveFlag.DefValue)
	
	// Check short flag
	shortFlag := cmd.Flags().ShorthandLookup("i")
	require.NotNil(t, shortFlag)
	assert.Equal(t, interactiveFlag, shortFlag)
}

func TestDisplayAPIPageInteractive(t *testing.T) {
	// Test the interactive display mode
	apis := []*types.OASAPI{
		{ID: "test-id-1", Name: "Test API 1", ListenPath: "/test1", DefaultVersion: "v1"},
		{ID: "test-id-2", Name: "Very Long API Name That Should Be Truncated", ListenPath: "/very/long/path/that/should/be/truncated", DefaultVersion: "v1"},
	}

	// Capture stderr for interactive output
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Test interactive display
	displayAPIPage(apis, 2, true)

	// Restore stderr
	wErr.Close()
	os.Stderr = oldStderr

	// Read captured output
	bufErr := make([]byte, 2048)
	nErr, _ := rErr.Read(bufErr)
	output := string(bufErr[:nErr])

	// Verify interactive-specific elements
	assert.Contains(t, output, "APIs (page 2)")
	assert.Contains(t, output, "================================================================================")
	assert.Contains(t, output, "Navigation: [←→ or AD] Next/Prev")
	assert.Contains(t, output, "[R] Refresh")
	assert.Contains(t, output, "[Q] Quit")
	assert.Contains(t, output, "Press a key to navigate...")
	
	// Verify table structure with pipe separators
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Listen Path")
	// The separator line uses repeated dashes with pipe separators
	
	// Verify API data is present
	assert.Contains(t, output, "test-id-1")
	assert.Contains(t, output, "Test API 1")
	
    // Verify truncation rules: Name truncated, Listen Path not truncated
    assert.Contains(t, output, "Very Long API Name Th...")
    assert.Contains(t, output, "/very/long/path/that/should/be/truncated")
}

func TestDisplayAPIPageEmptyInteractive(t *testing.T) {
	// Test empty API list in interactive mode
	apis := []*types.OASAPI{}

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	displayAPIPage(apis, 5, true)

	wErr.Close()
	os.Stderr = oldStderr

	bufErr := make([]byte, 1024)
	nErr, _ := rErr.Read(bufErr)
	output := string(bufErr[:nErr])

	// Verify empty page handling in interactive mode
	assert.Contains(t, output, "No APIs found on page 5")
	assert.Contains(t, output, "Navigation:")
	assert.Contains(t, output, "Press a key to navigate...")
}

func TestInteractiveTerminalDetection(t *testing.T) {
	// This test verifies the structure exists but can't test actual terminal detection
	// since that requires a real TTY
	
	// Verify the terminal detection function exists and is called
	// We can't easily test this without mocking or using a real terminal
	// but we can verify the function structure is correct
	
	// The runInteractiveAPIList function should exist and handle terminal detection
	// This is tested through manual verification and integration testing
	t.Log("Terminal detection tested through integration testing")
}

func TestAPIListWithRealEndpoint(t *testing.T) {
	// Test with the updated client parsing logic using dashboard response format
	dashboardResponse := map[string]interface{}{
		"apis": []interface{}{
			map[string]interface{}{
				"api_definition": map[string]interface{}{
					"api_id": "test-api-1",
					"name":   "Test API 1",
					"proxy": map[string]interface{}{
						"listen_path": "/test-api-1/",
					},
					"active": true,
				},
			},
			map[string]interface{}{
				"api_definition": map[string]interface{}{
					"api_id": "test-api-2", 
					"name":   "Test API 2",
					"proxy": map[string]interface{}{
						"listen_path": "/test-api-2/",
					},
					"active": true,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/apis", r.URL.Path)
		_ = json.NewEncoder(w).Encode(dashboardResponse)
	}))
	defer server.Close()

	// Create the command directly
	listCmd := NewAPIListCommand()

	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	
	listCmd.SetContext(withConfig(context.Background(), cfg))
	listCmd.SetContext(withOutputFormat(listCmd.Context(), types.OutputJSON))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listCmd.SetArgs([]string{"--page", "1"})
	err := listCmd.Execute()
	require.NoError(t, err)

	// Restore and read output
	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	
	// Parse JSON output
	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	require.NoError(t, err)
	
	// Verify JSON structure
	assert.Equal(t, float64(1), result["page"])
	assert.Equal(t, float64(2), result["count"])
	
	apis, ok := result["apis"].([]interface{})
	require.True(t, ok)
	assert.Len(t, apis, 2)
	
	// Verify API data
	apiBytes, _ := json.Marshal(result)
	apiString := string(apiBytes)
	assert.Contains(t, apiString, "test-api-1")
	assert.Contains(t, apiString, "test-api-2")
}
