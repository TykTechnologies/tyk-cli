package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// ---------------------------------------------------------------------------
// Test data: shared across all policy tests
// ---------------------------------------------------------------------------

// mockDashboardPolicy returns a DashboardPolicy JSON-encodable map
// Verifies: SYS-REQ-024
func mockDashboardPolicy(mid, wireID, name string, rate, per, quotaMax, quotaRenewalRate int64, tags []string, accessRights map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"_id":                  mid,
		"id":                   wireID,
		"name":                 name,
		"org_id":               "org",
		"rate":                 rate,
		"per":                  per,
		"quota_max":            quotaMax,
		"quota_renewal_rate":   quotaRenewalRate,
		"key_expires_in":       0,
		"tags":                 tags,
		"access_rights":        accessRights,
		"active":               true,
		"is_inactive":          false,
	}
}

// Verifies: SYS-REQ-024
func mockPolicyListResponse(policies []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"Data":       policies,
		"Pages":      1,
		"StatusCode": 200,
	}
}

// Verifies: SYS-REQ-001
func mockAPIListResponse() map[string]interface{} {
	return map[string]interface{}{
		"apis": []interface{}{
			map[string]interface{}{
				"api_definition": map[string]interface{}{
					"api_id": "a1b2c3d4e5f6",
					"name":   "users-api",
					"proxy": map[string]interface{}{
						"listen_path": "/users/",
					},
				},
			},
			map[string]interface{}{
				"api_definition": map[string]interface{}{
					"api_id": "g7h8i9j0k1l2",
					"name":   "orders-api",
					"proxy": map[string]interface{}{
						"listen_path": "/orders/",
					},
				},
			},
			map[string]interface{}{
				"api_definition": map[string]interface{}{
					"api_id": "m3n4o5p6q7r8",
					"name":   "payments-api",
					"proxy": map[string]interface{}{
						"listen_path": "/payments/",
					},
				},
			},
		},
	}
}

// Verifies: SYS-REQ-030
func goldPolicyAccessRights() map[string]interface{} {
	return map[string]interface{}{
		"a1b2c3d4e5f6": map[string]interface{}{
			"api_id":       "a1b2c3d4e5f6",
			"api_name":     "users-api",
			"versions":     []interface{}{"v1"},
			"allowed_urls": []interface{}{},
			"limit":        nil,
		},
		"g7h8i9j0k1l2": map[string]interface{}{
			"api_id":       "g7h8i9j0k1l2",
			"api_name":     "orders-api",
			"versions":     []interface{}{"v1", "v2"},
			"allowed_urls": []interface{}{},
			"limit":        nil,
		},
		"m3n4o5p6q7r8": map[string]interface{}{
			"api_id":       "m3n4o5p6q7r8",
			"api_name":     "payments-api",
			"versions":     []interface{}{"v1"},
			"allowed_urls": []interface{}{},
			"limit":        nil,
		},
	}
}

// Verifies: SYS-REQ-041
func createPolicyConfig(serverURL string) *types.Config {
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

// Verifies: SYS-REQ-028
func writeTempPolicyFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "policy.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

// validPlatinumPolicyYAML is the canonical test policy file content.
const validPlatinumPolicyYAML = `id: platinum
name: Platinum Plan
tags: [platinum, paid]
rateLimit:
  requests: 5000
  per: 1m
quota:
  limit: 500000
  period: 30d
keyTTL: 0
access:
  - name: users-api
    versions: [v1]
`

// ===========================================================================
// Walking Skeleton Tests (implement FIRST)
// ===========================================================================

// Verifies: SYS-REQ-024
func executePolicyListCmd(t *testing.T, serverURL string, outputFormat types.OutputFormat, extraArgs ...string) error {
	t.Helper()
	root := NewRootCommand("test", "commit", "time")
	listCmd, _, err := root.Find([]string{"policy", "list"})
	require.NoError(t, err)

	cfg := createPolicyConfig(serverURL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, outputFormat)
	listCmd.SetContext(ctx)

	if len(extraArgs) > 0 {
		listCmd.SetArgs(extraArgs)
		_ = listCmd.ParseFlags(extraArgs)
	}

	return listCmd.RunE(listCmd, []string{})
}

// Verifies: SYS-REQ-024
func TestPolicyList_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/portal/policies") {
			_ = json.NewEncoder(w).Encode(mockPolicyListResponse(nil))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := executePolicyListCmd(t, server.URL, types.OutputHuman)

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)

	require.NoError(t, err)
	assert.Contains(t, string(stderr), "No policies found")
}

// Verifies: SYS-REQ-024
func TestPolicyList_WithPolicies(t *testing.T) {
	policies := []map[string]interface{}{
		mockDashboardPolicy("507f1f77bcf86cd799439011", "gold", "Gold Plan", 1000, 60, 100000, 2592000,
			[]string{"gold", "paid"}, goldPolicyAccessRights()),
		mockDashboardPolicy("507f1f77bcf86cd799439012", "silver", "Silver Plan", 500, 60, 50000, 2592000,
			[]string{"silver"}, map[string]interface{}{
				"a1b2c3d4e5f6": map[string]interface{}{
					"api_id": "a1b2c3d4e5f6", "api_name": "users-api",
					"versions": []interface{}{"v1"},
				},
			}),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api/portal/policies") {
			_ = json.NewEncoder(w).Encode(mockPolicyListResponse(policies))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Capture stdout (table data goes to stdout)
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executePolicyListCmd(t, server.URL, types.OutputHuman)

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)
	output := string(stdout)
	// displayPolicyPage extracts friendly IDs from wire id field
	assert.Contains(t, output, "gold")
	assert.Contains(t, output, "Gold Plan")
	assert.Contains(t, output, "silver")
	assert.Contains(t, output, "Silver Plan")
}

// Verifies: SYS-REQ-026
func executePolicyApplyCmd(t *testing.T, serverURL string, filePath string) error {
	t.Helper()
	applyCmd := NewPolicyApplyCommand()

	cfg := createPolicyConfig(serverURL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputHuman)
	applyCmd.SetContext(ctx)

	applyCmd.SetArgs([]string{"-f", filePath})
	_ = applyCmd.ParseFlags([]string{"-f", filePath})

	return applyCmd.RunE(applyCmd, []string{})
}

// Verifies: SYS-REQ-026
func TestPolicyApply_Create_NameSelector(t *testing.T) {

	var capturedCreateBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Selector resolution: list APIs
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())

		// Resolve policy by ID — not found -> create path
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})

		// Create policy
		case r.Method == http.MethodPost && r.URL.Path == "/api/portal/policies":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &capturedCreateBody)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"Status": "success", "Message": "created", "Meta": "507f1f77bcf86cd799439099",
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.NoError(t, err)

	// Verify the wire format sent to Dashboard
	require.NotNil(t, capturedCreateBody)
	// _id should be empty/absent on create — Dashboard generates it
	mid, hasMID := capturedCreateBody["_id"]
	assert.True(t, !hasMID || mid == "", "_id should be empty or absent on create, got: %v", mid)
	// wire id should be the managed ID
	assert.Equal(t, "platinum", capturedCreateBody["id"])
	assert.Equal(t, "Platinum Plan", capturedCreateBody["name"])
	// Duration conversion: "1m" -> 60 seconds
	assert.EqualValues(t, 5000, capturedCreateBody["rate"])
	assert.EqualValues(t, 60, capturedCreateBody["per"])
	// Duration conversion: "30d" -> 2592000 seconds
	assert.EqualValues(t, 500000, capturedCreateBody["quota_max"])
	assert.EqualValues(t, 2592000, capturedCreateBody["quota_renewal_rate"])

	// Verify selector resolved: name "users-api" -> API ID in access_rights
	accessRights, ok := capturedCreateBody["access_rights"].(map[string]interface{})
	require.True(t, ok, "access_rights should be a map")
	_, hasUsersAPI := accessRights["a1b2c3d4e5f6"]
	assert.True(t, hasUsersAPI, "access_rights should contain resolved API ID a1b2c3d4e5f6")
}

// Verifies: SYS-REQ-026
func TestPolicyApply_Update_Idempotent(t *testing.T) {

	var capturedUpdateBody map[string]interface{}
	updateCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		// Selector resolution: list APIs
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())

		// Resolve policy by ID — found -> update path
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			existing := mockDashboardPolicy("507f1f77bcf86cd799439013", "platinum", "Platinum Plan", 5000, 60, 500000, 2592000,
				[]string{"platinum", "paid"}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(existing)

		// Update policy — PUT by id
		case r.Method == http.MethodPut && r.URL.Path == "/api/portal/policies/platinum":
			updateCalled = true
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &capturedUpdateBody)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"Status": "success", "Message": "updated", "Meta": "507f1f77bcf86cd799439013",
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Policy with updated rate limit (10000 instead of 5000)
	updatedYAML := strings.Replace(validPlatinumPolicyYAML, "requests: 5000", "requests: 10000", 1)
	policyFile := writeTempPolicyFile(t, updatedYAML)

	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.NoError(t, err)

	assert.True(t, updateCalled, "should have called PUT for existing policy")
	require.NotNil(t, capturedUpdateBody)
	assert.EqualValues(t, 10000, capturedUpdateBody["rate"])
}

// ===========================================================================
// Milestone 1: List + Get (focused scenarios)
// ===========================================================================

// Verifies: SYS-REQ-024
func TestPolicyList_JSONOutput(t *testing.T) {
	policies := []map[string]interface{}{
		mockDashboardPolicy("507f1f77bcf86cd799439011", "gold", "Gold Plan", 1000, 60, 100000, 2592000,
			[]string{"gold", "paid"}, goldPolicyAccessRights()),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(mockPolicyListResponse(policies))
	}))
	defer server.Close()

	// Capture stdout (JSON goes to stdout)
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executePolicyListCmd(t, server.URL, types.OutputJSON)

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)

	// Verify valid JSON with expected structure
	var result map[string]interface{}
	err = json.Unmarshal(stdout, &result)
	require.NoError(t, err, "output should be valid JSON")
	assert.NotNil(t, result["policies"])
	assert.Equal(t, float64(1), result["page"])
	assert.Equal(t, float64(1), result["count"])
}

// Verifies: SYS-REQ-024
func TestPolicyList_Pagination_EmptyPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2", r.URL.Query().Get("p"))
		_ = json.NewEncoder(w).Encode(mockPolicyListResponse(nil))
	}))
	defer server.Close()

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := executePolicyListCmd(t, server.URL, types.OutputHuman, "--page", "2")

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)

	require.NoError(t, err)
	assert.Contains(t, string(stderr), "No policies found")
}

// Verifies: SYS-REQ-024
func TestPolicyList_NetworkError(t *testing.T) {
	// Use a server that is immediately closed to trigger a network error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	err := executePolicyListCmd(t, server.URL, types.OutputHuman)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 1, exitErr.Code)
}

// Verifies: SYS-REQ-024
func TestPolicyCommand_Registration(t *testing.T) {
	root := NewRootCommand("test", "commit", "time")

	// Verify 'policy' appears in root subcommands
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "policy" {
			found = true
			// Verify subcommands of 'policy'
			subNames := make(map[string]bool)
			for _, sub := range cmd.Commands() {
				subNames[sub.Name()] = true
			}
			assert.True(t, subNames["list"], "'list' should be a subcommand of 'policy'")
			assert.True(t, subNames["get"], "'get' should be a subcommand of 'policy'")
			assert.True(t, subNames["apply"], "'apply' should be a subcommand of 'policy'")
			assert.True(t, subNames["delete"], "'delete' should be a subcommand of 'policy'")
		}
	}
	assert.True(t, found, "'policy' should be a subcommand of root")
}

// Verifies: SYS-REQ-025
func executePolicyGetCmd(t *testing.T, serverURL string, outputFormat types.OutputFormat, policyID string) error {
	t.Helper()
	root := NewRootCommand("test", "commit", "time")
	getCmd, _, err := root.Find([]string{"policy", "get"})
	require.NoError(t, err)

	cfg := createPolicyConfig(serverURL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, outputFormat)
	getCmd.SetContext(ctx)

	return getCmd.RunE(getCmd, []string{policyID})
}

// Verifies: SYS-REQ-025
func TestPolicyGet_Human(t *testing.T) {
	goldPolicy := mockDashboardPolicy("507f1f77bcf86cd799439011", "gold", "Gold Plan", 1000, 60, 100000, 2592000,
		[]string{"gold", "paid"}, goldPolicyAccessRights())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/gold":
			_ = json.NewEncoder(w).Encode(goldPolicy)
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Capture both stdout (YAML) and stderr (summary)
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	err := executePolicyGetCmd(t, server.URL, types.OutputHuman, "gold")

	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout, _ := io.ReadAll(rOut)
	stderr, _ := io.ReadAll(rErr)

	require.NoError(t, err)

	// stderr should contain human-readable summary
	stderrStr := string(stderr)
	assert.Contains(t, stderrStr, "Gold Plan")

	// Verify YAML is parseable
	var yamlResult map[string]interface{}
	err = yaml.Unmarshal(stdout, &yamlResult)
	require.NoError(t, err, "stdout should be valid YAML")
}

// Verifies: SYS-REQ-025
func TestPolicyGet_JSON(t *testing.T) {
	goldPolicy := mockDashboardPolicy("507f1f77bcf86cd799439011", "gold", "Gold Plan", 1000, 60, 100000, 2592000,
		[]string{"gold", "paid"}, goldPolicyAccessRights())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/gold":
			_ = json.NewEncoder(w).Encode(goldPolicy)
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executePolicyGetCmd(t, server.URL, types.OutputJSON, "gold")

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(stdout, &result)
	require.NoError(t, err, "output should be valid JSON")

	assert.Equal(t, "gold", result["id"])
	assert.Equal(t, "Gold Plan", result["name"])
}

// Verifies: SYS-REQ-035
func TestPolicyGet_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/nonexistent":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executePolicyGetCmd(t, server.URL, types.OutputHuman, "nonexistent")

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, exitErr.Message, "not found")
}

// ===========================================================================
// Milestone 2: Apply (focused scenarios)
// ===========================================================================

// Verifies: SYS-REQ-026
func TestPolicyApply_ListenPathSelector(t *testing.T) {

	var capturedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/path-test":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/portal/policies":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &capturedBody)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Meta": "path-test"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyYAML := `id: path-test
name: Path Test Policy
rateLimit:
  requests: 100
  per: 60
quota:
  limit: 10000
  period: 86400
keyTTL: 0
access:
  - listenPath: /orders/
    versions: [v1]
`
	policyFile := writeTempPolicyFile(t, policyYAML)

	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.NoError(t, err)

	// Verify listenPath "/orders/" resolved to "g7h8i9j0k1l2"
	accessRights, ok := capturedBody["access_rights"].(map[string]interface{})
	require.True(t, ok)
	_, hasOrdersAPI := accessRights["g7h8i9j0k1l2"]
	assert.True(t, hasOrdersAPI, "listenPath /orders/ should resolve to g7h8i9j0k1l2")
}

// Verifies: SYS-REQ-031
func TestPolicyApply_DurationConversion(t *testing.T) {

	var capturedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/dur-test":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/portal/policies":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &capturedBody)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Meta": "dur-test"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyYAML := `id: dur-test
name: Duration Test
rateLimit:
  requests: 1000
  per: 1m
quota:
  limit: 100000
  period: 30d
keyTTL: 24h
access:
  - name: users-api
    versions: [v1]
`
	policyFile := writeTempPolicyFile(t, policyYAML)

	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.NoError(t, err)

	assert.EqualValues(t, 60, capturedBody["per"], "1m should convert to 60 seconds")
	assert.EqualValues(t, 2592000, capturedBody["quota_renewal_rate"], "30d should convert to 2592000 seconds")
	assert.EqualValues(t, 86400, capturedBody["key_expires_in"], "24h should convert to 86400 seconds")
}

// Verifies: SYS-REQ-032
func TestPolicyApply_NameNotFound(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/apis" {
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	policyYAML := `id: typo-test
name: Typo Test
rateLimit:
  requests: 100
  per: 60
quota:
  limit: 10000
  period: 86400
keyTTL: 0
access:
  - name: inventori-api
    versions: [v1]
`
	policyFile := writeTempPolicyFile(t, policyYAML)

	err := executePolicyApplyCmd(t, server.URL, policyFile)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 2, exitErr.Code)
	assert.Contains(t, exitErr.Message, "no API found")
}

// Verifies: SYS-REQ-032
func TestPolicyApply_NameAmbiguous(t *testing.T) {

	// Mock server with two APIs named "api-service"
	ambiguousAPIList := map[string]interface{}{
		"apis": []interface{}{
			map[string]interface{}{
				"api_definition": map[string]interface{}{
					"api_id": "dup-1", "name": "api-service",
					"proxy": map[string]interface{}{"listen_path": "/svc-1/"},
				},
			},
			map[string]interface{}{
				"api_definition": map[string]interface{}{
					"api_id": "dup-2", "name": "api-service",
					"proxy": map[string]interface{}{"listen_path": "/svc-2/"},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/apis" {
			_ = json.NewEncoder(w).Encode(ambiguousAPIList)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	policyYAML := `id: ambig-test
name: Ambiguous Test
rateLimit:
  requests: 100
  per: 60
quota:
  limit: 10000
  period: 86400
keyTTL: 0
access:
  - name: api-service
    versions: [v1]
`
	policyFile := writeTempPolicyFile(t, policyYAML)

	err := executePolicyApplyCmd(t, server.URL, policyFile)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 2, exitErr.Code)
	assert.Contains(t, exitErr.Message, "ambiguous")
}

// Verifies: SYS-REQ-034
func TestPolicyApply_MissingID(t *testing.T) {

	policyYAML := `name: No ID Policy
rateLimit:
  requests: 100
  per: 60
quota:
  limit: 10000
  period: 86400
keyTTL: 0
access:
  - name: users-api
    versions: [v1]
`
	policyFile := writeTempPolicyFile(t, policyYAML)

	err := executePolicyApplyCmd(t, "http://unused", policyFile)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 2, exitErr.Code)
	assert.Contains(t, exitErr.Message, "id")
}

// Verifies: SYS-REQ-034
func TestPolicyApply_InvalidDuration(t *testing.T) {

	policyYAML := `id: bad-dur
name: Bad Duration
rateLimit:
  requests: 100
  per: abc
quota:
  limit: 10000
  period: 86400
keyTTL: 0
access:
  - name: users-api
    versions: [v1]
`
	policyFile := writeTempPolicyFile(t, policyYAML)

	err := executePolicyApplyCmd(t, "http://unused", policyFile)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 2, exitErr.Code)
	assert.Contains(t, exitErr.Message, "invalid duration")
}

// Verifies: SYS-REQ-034
func TestPolicyApply_FileNotFound(t *testing.T) {
	err := executePolicyApplyCmd(t, "http://unused", "/nonexistent/policy.yaml")

	require.Error(t, err)
}

// ===========================================================================
// Milestone 3: Delete + Init
// ===========================================================================

// Verifies: SYS-REQ-027
func executePolicyDeleteCmd(t *testing.T, serverURL string, outputFormat types.OutputFormat, policyID string, yes bool) error {
	t.Helper()
	deleteCmd := NewPolicyDeleteCommand()

	cfg := createPolicyConfig(serverURL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, outputFormat)
	deleteCmd.SetContext(ctx)

	cmdArgs := []string{policyID}
	if yes {
		cmdArgs = append(cmdArgs, "--yes")
	}
	deleteCmd.SetArgs(cmdArgs)
	_ = deleteCmd.ParseFlags(cmdArgs)

	return deleteCmd.RunE(deleteCmd, []string{policyID})
}

// Verifies: SYS-REQ-027
func TestPolicyDelete_WithYes(t *testing.T) {
	deleteCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/free-tier":
			p := mockDashboardPolicy("507f1f77bcf86cd799439020", "free-tier", "Free Plan", 100, 60, 10000, 86400,
				[]string{"free"}, map[string]interface{}{
					"a1b2c3d4e5f6": map[string]interface{}{"api_id": "a1b2c3d4e5f6"},
				})
			_ = json.NewEncoder(w).Encode(p)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/portal/policies/free-tier":
			deleteCalled = true
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Message": "deleted"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Capture stderr for confirmation message
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "free-tier", true)

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)

	require.NoError(t, err)
	assert.True(t, deleteCalled, "DELETE should have been called")
	assert.Contains(t, string(stderr), "Free Plan", "stderr should mention policy name")
	assert.Contains(t, string(stderr), "deleted", "stderr should confirm deletion")
}

// Verifies: SYS-REQ-035
func TestPolicyDelete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/nonexistent":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "nonexistent", true)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Equal(t, 3, exitErr.Code)
	assert.Contains(t, exitErr.Message, "not found")
}

// Verifies: SYS-REQ-027
func TestPolicyDelete_WithYes_JSON(t *testing.T) {
	deleteCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/free-tier":
			p := mockDashboardPolicy("507f1f77bcf86cd799439020", "free-tier", "Free Plan", 100, 60, 10000, 86400,
				[]string{"free"}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/portal/policies/free-tier":
			deleteCalled = true
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Message": "deleted"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Capture stdout for JSON output
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err := executePolicyDeleteCmd(t, server.URL, types.OutputJSON, "free-tier", true)

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)

	require.NoError(t, err)
	assert.True(t, deleteCalled, "DELETE should have been called")

	// Verify structured JSON output
	var result map[string]interface{}
	err = json.Unmarshal(stdout, &result)
	require.NoError(t, err, "output should be valid JSON")
	assert.Equal(t, "free-tier", result["policy_id"])
	assert.Equal(t, "deleted", result["operation"])
	assert.Equal(t, true, result["success"])
}

// Verifies: SYS-REQ-028
func executePolicyInitCmd(t *testing.T, dir string, id string, name string) error {
	t.Helper()
	initCmd := NewPolicyInitCommand()

	// No config needed -- init is offline
	ctx := context.Background()
	initCmd.SetContext(ctx)

	args := []string{"--id", id, "--name", name, "--dir", dir}
	initCmd.SetArgs(args)
	if err := initCmd.ParseFlags(args); err != nil {
		return err
	}

	return initCmd.RunE(initCmd, []string{})
}

// Verifies: SYS-REQ-028
func TestPolicyInit_NewFile(t *testing.T) {
	tmpDir := t.TempDir()

	err := executePolicyInitCmd(t, tmpDir, "my-policy", "My Policy")
	require.NoError(t, err)

	// Verify file was created at policies/{id}.yaml inside the dir
	outPath := filepath.Join(tmpDir, "policies", "my-policy.yaml")
	data, err := os.ReadFile(outPath)
	require.NoError(t, err, "scaffold file should exist at policies/{id}.yaml")

	// Parse and validate the scaffold YAML
	var pf types.PolicyFile
	err = yaml.Unmarshal(data, &pf)
	require.NoError(t, err, "scaffold should be valid YAML")

	// Verify schema fields
	assert.Equal(t, "my-policy", pf.ID)
	assert.Equal(t, "My Policy", pf.Name)

	// Verify sensible defaults
	require.NotNil(t, pf.RateLimit, "scaffold should have default rateLimit")
	assert.Equal(t, int64(1000), pf.RateLimit.Requests)
	assert.Equal(t, types.Duration("1m"), pf.RateLimit.Per)
	require.NotNil(t, pf.Quota, "scaffold should have default quota")
	assert.Equal(t, int64(100000), pf.Quota.Limit)
	assert.Equal(t, types.Duration("30d"), pf.Quota.Period)
	assert.Equal(t, types.Duration("0"), pf.KeyTTL)
	require.Len(t, pf.Access, 1, "scaffold should have one placeholder access entry")
	assert.Equal(t, "your-api-name", pf.Access[0].Name)
	assert.Equal(t, []string{"Default"}, pf.Access[0].Versions)
}

// Verifies: SYS-REQ-028
func TestPolicyInit_FileExistsNoOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the file that init would write to
	policiesDir := filepath.Join(tmpDir, "policies")
	require.NoError(t, os.MkdirAll(policiesDir, 0755))
	existingPath := filepath.Join(policiesDir, "existing.yaml")
	require.NoError(t, os.WriteFile(existingPath, []byte("original content"), 0644))

	err := executePolicyInitCmd(t, tmpDir, "existing", "Existing Policy")

	require.Error(t, err, "init should error when file already exists")
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "should return ExitError")
	assert.Contains(t, exitErr.Message, "already exists")

	// Verify original content untouched
	data, _ := os.ReadFile(existingPath)
	assert.Equal(t, "original content", string(data))
}

// Verifies: SYS-REQ-028
func TestPolicyInit_Registration(t *testing.T) {
	root := NewRootCommand("test", "commit", "time")

	// Find init under policy
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "policy" {
			for _, sub := range cmd.Commands() {
				if sub.Name() == "init" {
					found = true
				}
			}
		}
	}
	assert.True(t, found, "'init' should be a subcommand of 'policy'")
}

// ===========================================================================
// Full Integration Walking Skeleton
// ===========================================================================

// ---------------------------------------------------------------------------
// MC/DC coverage for runPolicyList / runPolicyGet / runPolicyApply /
// runPolicyDelete / runPolicyInit / readPolicyFile / buildResolveRequests /
// resolveFriendlyID / isNotFoundError / displayPolicyPage.
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-024
// brokenPolicyConfig returns a config whose default_environment does not exist
// so client.NewClient fails.
func brokenPolicyConfig() *types.Config {
	return &types.Config{
		DefaultEnvironment: "ghost",
		Environments: map[string]*types.Environment{
			"other": {Name: "other", DashboardURL: "http://x:1", AuthToken: "t", OrgID: "o"},
		},
	}
}

// Verifies: SYS-REQ-024
// TestRunPolicyList_NewClientFails covers L68 err!=nil from client.NewClient.
func TestRunPolicyList_NewClientFails(t *testing.T) {
	cmd := NewPolicyListCommand()
	cmd.SetContext(withConfig(context.Background(), brokenPolicyConfig()))
	cmd.SetArgs([]string{})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// Verifies: SYS-REQ-025
// TestRunPolicyGet_ListAPIsFailureNonFatal covers L156 err!=nil branch from
// the /api/apis call. The policy get function logs and continues even when
// API list fetch fails.
func TestRunPolicyGet_ListAPIsFailureNonFatal(t *testing.T) {
	pol := mockDashboardPolicy("507f1f77bcf86cd799439011", "gold", "Gold Plan", 1000, 60, 100000, 2592000,
		[]string{}, map[string]interface{}{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/portal/policies/gold":
			_ = json.NewEncoder(w).Encode(pol)
		case r.URL.Path == "/api/apis":
			// Make the API list fetch fail.
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executePolicyGetCmd(t, server.URL, types.OutputJSON, "gold")
	require.NoError(t, err, "API list failure should be non-fatal in policy get")
}

// Verifies: SYS-REQ-025
// TestRunPolicyGet_NewClientFails covers L131 err!=nil from client.NewClient.
func TestRunPolicyGet_NewClientFails(t *testing.T) {
	cmd := NewPolicyGetCommand()
	cmd.SetContext(withConfig(context.Background(), brokenPolicyConfig()))
	cmd.SetArgs([]string{"x"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_FetchAPIList_NonClassified covers L241 cls!=nil=F branch
// (404 status from /api/apis is not in {401,403,429,5xx}).
func TestRunPolicyApply_FetchAPIList_NonClassified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 400 is not in the classified set.
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
	}))
	defer server.Close()

	tmpFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, tmpFile)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 1, exitErr.Code, "non-classified errors map to exit 1")
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_UpdateNonClassified covers L276 cls!=nil=F branch.
func TestRunPolicyApply_UpdateNonClassified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			p := mockDashboardPolicy("507f1f77bcf86cd799439013", "platinum", "Platinum Plan", 5000, 60, 500000, 2592000,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.Method == http.MethodPut && r.URL.Path == "/api/portal/policies/platinum":
			// 400 is not classified
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tmpFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, tmpFile)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 1, exitErr.Code)
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_CreateNonClassified covers L289 cls!=nil=F branch.
func TestRunPolicyApply_CreateNonClassified(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/portal/policies":
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tmpFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, tmpFile)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 1, exitErr.Code)
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_NewClientFails covers L231 err!=nil from client.NewClient.
func TestRunPolicyApply_NewClientFails(t *testing.T) {
	tmpFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	cmd := NewPolicyApplyCommand()
	cmd.SetContext(withConfig(context.Background(), brokenPolicyConfig()))
	cmd.SetArgs([]string{"-f", tmpFile})
	_ = cmd.ParseFlags([]string{"-f", tmpFile})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_NewClientFails covers L328 err!=nil from client.NewClient.
func TestRunPolicyDelete_NewClientFails(t *testing.T) {
	cmd := NewPolicyDeleteCommand()
	cmd.SetContext(withConfig(context.Background(), brokenPolicyConfig()))
	cmd.SetArgs([]string{"x", "--yes"})
	_ = cmd.ParseFlags([]string{"x", "--yes"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
}

// Verifies: SYS-REQ-024
// TestRunPolicyList_ConfigNil covers L62 config==nil=T.
func TestRunPolicyList_ConfigNil(t *testing.T) {
	cmd := NewPolicyListCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// Verifies: SYS-REQ-024
// TestRunPolicyList_PageZeroDefaults covers L56 page<=0=T branch.
func TestRunPolicyList_PageZeroDefaults(t *testing.T) {
	gotPage := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPage = r.URL.Query().Get("p")
		_ = json.NewEncoder(w).Encode(mockPolicyListResponse(nil))
	}))
	defer server.Close()

	cmd := NewPolicyListCommand()
	cfg := createPolicyConfig(server.URL)
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputHuman))
	cmd.SetArgs([]string{"--page", "0"})
	_ = cmd.ParseFlags([]string{"--page", "0"})
	require.NoError(t, cmd.RunE(cmd, []string{}))
	assert.Equal(t, "1", gotPage)
}

// Verifies: SYS-REQ-025
// TestRunPolicyGet_ConfigNil covers L125 config==nil=T.
func TestRunPolicyGet_ConfigNil(t *testing.T) {
	cmd := NewPolicyGetCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"x"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// Verifies: SYS-REQ-025
// TestRunPolicyGet_HumanWithTags covers L174 len(pf.Tags)>0=T branch.
// Already covered by TestPolicyGet_Human, but include the targeted assertion.
func TestRunPolicyGet_HumanWithTags(t *testing.T) {
	pol := mockDashboardPolicy("507f1f77bcf86cd799439011", "gold", "Gold Plan", 1000, 60, 100000, 2592000,
		[]string{"gold", "paid"}, goldPolicyAccessRights())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/portal/policies/gold":
			_ = json.NewEncoder(w).Encode(pol)
		case "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr
	_ = executePolicyGetCmd(t, server.URL, types.OutputHuman, "gold")
	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)
	assert.Contains(t, string(stderr), "Tags:")
}

// Verifies: SYS-REQ-025
// TestRunPolicyGet_HumanNoTags covers L174 len(pf.Tags)>0=F branch.
func TestRunPolicyGet_HumanNoTags(t *testing.T) {
	pol := mockDashboardPolicy("507f1f77bcf86cd799439011", "no-tags", "No Tags", 100, 60, 1000, 86400,
		[]string{}, map[string]interface{}{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/portal/policies/no-tags":
			_ = json.NewEncoder(w).Encode(pol)
		case "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr
	_ = executePolicyGetCmd(t, server.URL, types.OutputHuman, "no-tags")
	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)
	assert.NotContains(t, string(stderr), "Tags:")
}

// Verifies: SYS-REQ-025
// TestRunPolicyGet_NonClassifiedError covers L145 cls!=nil=F branch (a 400
// from GetPolicy is not classified, so cls==nil and we fall through to wrap).
func TestRunPolicyGet_NonClassifiedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad input"})
	}))
	defer server.Close()

	err := executePolicyGetCmd(t, server.URL, types.OutputHuman, "x")
	require.Error(t, err)
}

// Verifies: SYS-REQ-025
// TestRunPolicyGet_DashboardError covers L145 classifyDashboardError cls!=nil path.
func TestRunPolicyGet_DashboardError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 401, "message": "bad token"})
	}))
	defer server.Close()

	err := executePolicyGetCmd(t, server.URL, types.OutputHuman, "x")
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitAuthFailed), exitErr.Code)
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_ConfigNil covers L221 config==nil=T.
func TestRunPolicyApply_ConfigNil(t *testing.T) {
	tmpFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	cmd := NewPolicyApplyCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"-f", tmpFile})
	_ = cmd.ParseFlags([]string{"-f", tmpFile})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_ResolveError covers L265 resolveErr!=nil=T path (the
// policy GET call returns a non-404, non-success error).
func TestRunPolicyApply_ResolveErrorBubble(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.URL.Path == "/api/portal/policies/platinum":
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "boom"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.Error(t, err)
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_FetchAPIListAuthError covers L241 cls!=nil branch.
func TestRunPolicyApply_FetchAPIListAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 401, "message": "no auth"})
	}))
	defer server.Close()

	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitAuthFailed), exitErr.Code)
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_UpdateServerError covers L276 classifyDashboardError path
// on update with a 500.
func TestRunPolicyApply_UpdateServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			p := mockDashboardPolicy("507f1f77bcf86cd799439013", "platinum", "Platinum Plan", 5000, 60, 500000, 2592000,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.Method == http.MethodPut && r.URL.Path == "/api/portal/policies/platinum":
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "boom"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitServerError), exitErr.Code)
}

// Verifies: SYS-REQ-026
// TestRunPolicyApply_CreateServerError covers the create-path classifyDashboardError.
func TestRunPolicyApply_CreateServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/portal/policies":
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "boom"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, policyFile)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitServerError), exitErr.Code)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_ConfigNil covers L322 config==nil=T.
func TestRunPolicyDelete_ConfigNil(t *testing.T) {
	cmd := NewPolicyDeleteCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"x", "--yes"})
	_ = cmd.ParseFlags([]string{"x", "--yes"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_PromptCancelled covers L353 strings.ToLower(response) !=
// "y" && != "yes" decision with response="n".
func TestRunPolicyDelete_PromptCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodGet {
			p := mockDashboardPolicy("507f1f77bcf86cd799439020", "free-tier", "Free Plan", 100, 60, 10000, 86400,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	var err error
	withStdin(t, "n\n", func() {
		err = executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "free-tier", false)
	})
	require.NoError(t, err)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_PromptConfirmedY covers L353 response=="y" branch.
func TestRunPolicyDelete_PromptConfirmedY(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodGet:
			p := mockDashboardPolicy("507f1f77bcf86cd799439020", "free-tier", "Free Plan", 100, 60, 10000, 86400,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodDelete:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Message": "deleted"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var err error
	withStdin(t, "y\n", func() {
		err = executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "free-tier", false)
	})
	require.NoError(t, err)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_PromptConfirmedYes covers L353 response=="yes" branch.
func TestRunPolicyDelete_PromptConfirmedYes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodGet:
			p := mockDashboardPolicy("507f1f77bcf86cd799439020", "free-tier", "Free Plan", 100, 60, 10000, 86400,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodDelete:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Message": "deleted"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var err error
	withStdin(t, "yes\n", func() {
		err = executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "free-tier", false)
	})
	require.NoError(t, err)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_ResolveNonClassifiedError covers L339 cls!=nil=F branch
// (non-classified error during resolveFriendlyID).
func TestRunPolicyDelete_ResolveNonClassifiedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
	}))
	defer server.Close()

	err := executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "x", true)
	require.Error(t, err)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_DeleteNonClassifiedError covers L361 cls!=nil=F branch.
func TestRunPolicyDelete_DeleteNonClassifiedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodGet:
			p := mockDashboardPolicy("507f1f77bcf86cd799439020", "free-tier", "Free Plan", 100, 60, 10000, 86400,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "bad"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "free-tier", true)
	require.Error(t, err)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_ResolveError covers L339 cls!=nil branch (auth failure
// during resolveFriendlyID).
func TestRunPolicyDelete_ResolveAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 401, "message": "no"})
	}))
	defer server.Close()

	err := executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "x", true)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitAuthFailed), exitErr.Code)
}

// Verifies: SYS-REQ-027
// TestRunPolicyDelete_DeleteError covers L361 classifyDashboardError on delete.
func TestRunPolicyDelete_DeleteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodGet:
			p := mockDashboardPolicy("507f1f77bcf86cd799439020", "free-tier", "Free Plan", 100, 60, 10000, 86400,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.URL.Path == "/api/portal/policies/free-tier" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "boom"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "free-tier", true)
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitServerError), exitErr.Code)
}

// Verifies: SYS-REQ-028
// TestRunPolicyInit_WriteFails covers L463 err!=nil from os.WriteFile by
// making the target path a directory.
func TestRunPolicyInit_WriteFails(t *testing.T) {
	tmpDir := t.TempDir()
	policiesDir := filepath.Join(tmpDir, "policies")
	require.NoError(t, os.MkdirAll(policiesDir, 0o755))
	// Pre-create a directory named "x.yaml" so WriteFile cannot write to it.
	require.NoError(t, os.MkdirAll(filepath.Join(policiesDir, "y.yaml"), 0o755))

	cmd := NewPolicyInitCommand()
	cmd.SetArgs([]string{"--id", "y", "--name", "Y", "--dir", tmpDir})
	_ = cmd.ParseFlags([]string{"--id", "y", "--name", "Y", "--dir", tmpDir})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// Verifies: SYS-REQ-028
// TestRunPolicyInit_UnwritableDir covers L459 err!=nil from os.MkdirAll
// (failure to create the policies subdir because its parent is a regular file
// rather than a directory).
func TestRunPolicyInit_UnwritableDir(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a regular file where the "policies" directory would normally be.
	conflict := filepath.Join(tmpDir, "policies")
	require.NoError(t, os.WriteFile(conflict, []byte("not a dir"), 0o600))

	cmd := NewPolicyInitCommand()
	cmd.SetArgs([]string{"--id", "x", "--name", "X", "--dir", tmpDir})
	_ = cmd.ParseFlags([]string{"--id", "x", "--name", "X", "--dir", tmpDir})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// Verifies: SYS-REQ-028
// TestRunPolicyInit_MissingID covers L416 id==""=T.
func TestRunPolicyInit_MissingID(t *testing.T) {
	cmd := NewPolicyInitCommand()
	cmd.SetArgs([]string{"--name", "x", "--dir", "."})
	_ = cmd.ParseFlags([]string{"--name", "x", "--dir", "."})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitBadArgs), exitErr.Code)
}

// Verifies: SYS-REQ-028
// TestRunPolicyInit_MissingName covers L419 name==""=T.
func TestRunPolicyInit_MissingName(t *testing.T) {
	cmd := NewPolicyInitCommand()
	cmd.SetArgs([]string{"--id", "x", "--dir", "."})
	_ = cmd.ParseFlags([]string{"--id", "x", "--dir", "."})
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitBadArgs), exitErr.Code)
}

// Verifies: SYS-REQ-029
// TestReadPolicyFile_StdinHappy covers L476 filePath=="-"=T branch.
func TestReadPolicyFile_StdinHappy(t *testing.T) {
	yamlBody := validPlatinumPolicyYAML
	withStdin(t, yamlBody, func() {
		pf, err := readPolicyFile("-")
		require.NoError(t, err)
		assert.Equal(t, "platinum", pf.ID)
	})
}

// Verifies: SYS-REQ-029
// TestReadPolicyFile_StdinEmpty covers L481 len(data)==0=T.
func TestReadPolicyFile_StdinEmpty(t *testing.T) {
	withStdin(t, "", func() {
		_, err := readPolicyFile("-")
		require.Error(t, err)
		exitErr, ok := err.(*ExitError)
		require.True(t, ok)
		assert.Equal(t, int(types.ExitBadArgs), exitErr.Code)
	})
}

// Verifies: SYS-REQ-029
// TestReadPolicyFile_BadYAML covers L492 yaml.Unmarshal err!=nil branch.
func TestReadPolicyFile_BadYAML(t *testing.T) {
	tmpFile := writeTempPolicyFile(t, "::not valid yaml")
	_, err := readPolicyFile(tmpFile)
	require.Error(t, err)
}

// Verifies: SYS-REQ-032
// TestBuildResolveRequests_AllSelectorKinds drives every branch of the switch.
func TestBuildResolveRequests_AllSelectorKinds(t *testing.T) {
	entries := []types.AccessEntry{
		{ID: "abc"},
		{Name: "users-api"},
		{ListenPath: "/users/"},
		{Tags: []string{"premium", "paid"}},
		{}, // empty selector: takes default path (no match)
	}
	reqs := buildResolveRequests(entries)
	require.Len(t, reqs, 5)
	assert.Equal(t, "id", reqs[0].SelectorType)
	assert.Equal(t, "name", reqs[1].SelectorType)
	assert.Equal(t, "listenPath", reqs[2].SelectorType)
	assert.Equal(t, "tags", reqs[3].SelectorType)
	assert.Equal(t, "", reqs[4].SelectorType)
}

// Verifies: SYS-REQ-015
// TestIsNotFoundError_BothBranches covers L554 ok && er.Status==404 branch
// and the substring fallback.
func TestIsNotFoundError_BothBranches(t *testing.T) {
	// Typed error with status 404 -> true.
	er := &types.ErrorResponse{Status: 404, Message: "x"}
	assert.True(t, isNotFoundError(er))

	// Typed error with another status -> falls through, message no match -> false.
	er2 := &types.ErrorResponse{Status: 500, Message: "boom"}
	assert.False(t, isNotFoundError(er2))

	// Untyped error with "404" -> true.
	assert.True(t, isNotFoundError(fmt.Errorf("HTTP 404")))

	// Untyped error with "not found" -> true.
	assert.True(t, isNotFoundError(fmt.Errorf("api was Not Found")))

	// Untyped error with no signal -> false.
	assert.False(t, isNotFoundError(fmt.Errorf("something else")))
}

// Verifies: SYS-REQ-024
// TestDisplayPolicyPage_DisplayIDEmpty covers L586 displayID==""=T branch.
func TestDisplayPolicyPage_DisplayIDEmpty(t *testing.T) {
	policies := []types.DashboardPolicy{
		{ID: "", MID: "mid-fallback", Name: "Unmanaged Policy"},
	}
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	displayPolicyPage(policies, 1)
	wOut.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(rOut)
	assert.Contains(t, string(out), "mid-fallback")
}

// Verifies: SYS-REQ-024
func TestPolicyIntegration_FullLifecycle(t *testing.T) {
	// This test exercises: list empty -> apply new -> list shows policy -> get returns CLI schema -> delete removes
	// The mock Dashboard stores policies keyed by id (e.g., "platinum").
	policyStore := map[string]map[string]interface{}{} // keyed by id (e.g., "platinum")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())

		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies":
			policies := make([]map[string]interface{}, 0)
			for _, p := range policyStore {
				policies = append(policies, p)
			}
			_ = json.NewEncoder(w).Encode(mockPolicyListResponse(policies))

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/portal/policies/"):
			lookupID := strings.TrimPrefix(r.URL.Path, "/api/portal/policies/")
			if p, ok := policyStore[lookupID]; ok {
				_ = json.NewEncoder(w).Encode(p)
			} else {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
			}

		case r.Method == http.MethodPost && r.URL.Path == "/api/portal/policies":
			body, _ := io.ReadAll(r.Body)
			var created map[string]interface{}
			_ = json.Unmarshal(body, &created)
			created["_id"] = "507f1f77bcf86cd799439099"
			id, _ := created["id"].(string)
			policyStore[id] = created
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Message": "created", "Meta": "507f1f77bcf86cd799439099"})

		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/portal/policies/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/portal/policies/")
			body, _ := io.ReadAll(r.Body)
			var updated map[string]interface{}
			_ = json.Unmarshal(body, &updated)
			policyStore[id] = updated
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Message": "updated"})

		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/portal/policies/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/portal/policies/")
			delete(policyStore, id)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Status": "success", "Message": "deleted"})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Step 1: List should be empty
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	err := executePolicyListCmd(t, server.URL, types.OutputHuman)

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ := io.ReadAll(rErr)
	require.NoError(t, err)
	assert.Contains(t, string(stderr), "No policies found")

	// Step 2: Apply a new policy
	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err = executePolicyApplyCmd(t, server.URL, policyFile)
	require.NoError(t, err, "apply should succeed")

	// Step 3: List should now show the policy
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err = executePolicyListCmd(t, server.URL, types.OutputJSON)

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ := io.ReadAll(rOut)
	require.NoError(t, err)

	var listResult map[string]interface{}
	require.NoError(t, json.Unmarshal(stdout, &listResult))
	assert.Equal(t, float64(1), listResult["count"], "list should show 1 policy after apply")

	// Step 4: Get should return CLI schema
	oldStdout = os.Stdout
	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut

	err = executePolicyGetCmd(t, server.URL, types.OutputJSON, "platinum")

	wOut.Close()
	os.Stdout = oldStdout
	stdout, _ = io.ReadAll(rOut)
	require.NoError(t, err)

	var getResult map[string]interface{}
	require.NoError(t, json.Unmarshal(stdout, &getResult))
	assert.Equal(t, "platinum", getResult["id"])
	assert.Equal(t, "Platinum Plan", getResult["name"])

	// Step 5: Delete the policy
	err = executePolicyDeleteCmd(t, server.URL, types.OutputHuman, "platinum", true)
	require.NoError(t, err, "delete should succeed")

	// Step 6: List should be empty again
	oldStderr = os.Stderr
	rErr, wErr, _ = os.Pipe()
	os.Stderr = wErr

	err = executePolicyListCmd(t, server.URL, types.OutputHuman)

	wErr.Close()
	os.Stderr = oldStderr
	stderr, _ = io.ReadAll(rErr)
	require.NoError(t, err)
	assert.Contains(t, string(stderr), "No policies found")
}
