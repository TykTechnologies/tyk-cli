package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Verifies: SYS-REQ-003
func TestGenerateOASForCreate(t *testing.T) {
	tests := []struct {
		name         string
		apiName      string
		description  string
		version      string
		upstreamURL  string
		listenPath   string
		customDomain string
		checkResult  func(t *testing.T, result map[string]interface{})
	}{
		{
			name:        "basic API creation",
			apiName:     "User Service",
			description: "User management API",
			version:     "v1",
			upstreamURL: "https://users.api.com",
			listenPath:  "/user-service/",
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Check basic OAS structure
				assert.Equal(t, "3.0.0", result["openapi"])
				
				info, ok := result["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "User Service", info["title"])
				assert.Equal(t, "User management API", info["description"])
				assert.Equal(t, "v1", info["version"])
				
				servers, ok := result["servers"].([]interface{})
				require.True(t, ok)
				require.Len(t, servers, 1)
				
				server, ok := servers[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "https://users.api.com", server["url"])
				
				// Check Tyk extensions
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				extInfo, ok := tykExt["info"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "User Service", extInfo["name"])
				
				state, ok := extInfo["state"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, state["active"])
				
				upstream, ok := tykExt["upstream"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "https://users.api.com", upstream["url"])
				
				server, ok = tykExt["server"].(map[string]interface{})
				require.True(t, ok)
				
				listenPath, ok := server["listenPath"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "/user-service/", listenPath["value"])
				assert.Equal(t, true, listenPath["strip"])
			},
		},
		{
			name:         "API with custom domain",
			apiName:      "Payment API",
			description:  "Payment processing API",
			version:      "v2",
			upstreamURL:  "https://payments.internal",
			listenPath:   "/payments/v2/",
			customDomain: "api.company.com",
			checkResult: func(t *testing.T, result map[string]interface{}) {
				// Check custom domain in Tyk extensions
				tykExt, ok := result["x-tyk-api-gateway"].(map[string]interface{})
				require.True(t, ok)
				
				server, ok := tykExt["server"].(map[string]interface{})
				require.True(t, ok)
				
				customDomain, ok := server["customDomain"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, customDomain["enabled"])
				assert.Equal(t, "api.company.com", customDomain["name"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateOASForCreate(
				tt.apiName,
				tt.description,
				tt.version,
				tt.upstreamURL,
				tt.listenPath,
				tt.customDomain,
			)
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// Verifies: SYS-REQ-003
func TestNewAPICreateCommand(t *testing.T) {
	cmd := NewAPICreateCommand()

	assert.Equal(t, "create", cmd.Use)
	assert.Equal(t, "Create a new API from scratch", cmd.Short)
	assert.Contains(t, cmd.Long, "Create a new OAS API from scratch")

	// Check required flags
	nameFlag := cmd.Flags().Lookup("name")
	require.NotNil(t, nameFlag)

	upstreamFlag := cmd.Flags().Lookup("upstream-url")
	require.NotNil(t, upstreamFlag)

	// Check optional flags
	listenPathFlag := cmd.Flags().Lookup("listen-path")
	require.NotNil(t, listenPathFlag)

	versionFlag := cmd.Flags().Lookup("version-name")
	require.NotNil(t, versionFlag)
	assert.Equal(t, "v1", versionFlag.DefValue)

	customDomainFlag := cmd.Flags().Lookup("custom-domain")
	require.NotNil(t, customDomainFlag)

	descriptionFlag := cmd.Flags().Lookup("description")
	require.NotNil(t, descriptionFlag)
}

// ---------------------------------------------------------------------------
// MC/DC coverage for runAPICreate branches
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-003
// executeAPICreate builds and runs a create command against a test server.
// It returns the resulting error from RunE.
func executeAPICreate(t *testing.T, serverURL string, format types.OutputFormat, args []string) error {
	t.Helper()
	cmd := NewAPICreateCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: serverURL, AuthToken: "tok", OrgID: "org"},
		},
	}
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, format)
	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	return cmd.RunE(cmd, []string{})
}

// Verifies: SYS-REQ-003
// TestRunAPICreate_AutoListenPathAndDescription covers L1375 listenPath==""=T
// and L1380 description==""=T branches, plus L1418 outputFormat==OutputJSON=T
// (with JSON output).
func TestRunAPICreate_AutoListenPathAndDescription_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/apis/oas" {
			_ = json.NewEncoder(w).Encode(mockCreateAPIResponse())
			return
		}
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/apis/oas/") {
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := executeAPICreate(t, server.URL, types.OutputJSON, []string{
		"--name", "Auto API", "--upstream-url", "http://up.example.com",
	})
	wOut.Close()
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// Verifies: SYS-REQ-003
// TestRunAPICreate_AllFlagsSet covers L1375 listenPath==""=F and L1380
// description==""=F branches.
func TestRunAPICreate_AllFlagsSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/apis/oas" {
			_ = json.NewEncoder(w).Encode(mockCreateAPIResponse())
			return
		}
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/apis/oas/") {
			_ = json.NewEncoder(w).Encode(mockOASAPIResponse())
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	err := executeAPICreate(t, server.URL, types.OutputHuman, []string{
		"--name", "Full API",
		"--upstream-url", "http://up.example.com",
		"--listen-path", "/full/",
		"--description", "Custom desc",
		"--custom-domain", "api.example.com",
		"--version-name", "v2",
	})
	require.NoError(t, err)
}

// Verifies: SYS-REQ-003
// TestRunAPICreate_NewClientFails covers L1398 err!=nil from client.NewClient.
func TestRunAPICreate_NewClientFails(t *testing.T) {
	cmd := NewAPICreateCommand()
	cmd.SetContext(withConfig(context.Background(), brokenConfig()))
	args := []string{"--name", "X", "--upstream-url", "http://up"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// Verifies: SYS-REQ-003
// TestRunAPICreate_ConfigNil covers L1386 config==nil=T branch.
func TestRunAPICreate_ConfigNil(t *testing.T) {
	cmd := NewAPICreateCommand()
	cmd.SetContext(context.Background())
	args := []string{"--name", "X", "--upstream-url", "http://up"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// Verifies: SYS-REQ-003
// TestRunAPICreate_ServerError covers L1412 fmt.Errorf wrap for non-conflict
// dashboard errors.
func TestRunAPICreate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "boom"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	err := executeAPICreate(t, server.URL, types.OutputHuman, []string{
		"--name", "X", "--upstream-url", "http://up",
	})
	require.Error(t, err)
}