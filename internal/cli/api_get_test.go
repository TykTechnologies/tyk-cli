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

// reqproof:req REQ-API-002
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

// reqproof:req REQ-API-002
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

// reqproof:req REQ-API-002
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

// reqproof:req REQ-API-002
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

// reqproof:req REQ-API-002
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

// reqproof:req REQ-API-002
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

// ---------------------------------------------------------------------------
// MC/DC coverage for outputAPIAsHuman / outputAPIAsJSON
// ---------------------------------------------------------------------------

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_NilAPI covers L669 api==nil branch.
func TestOutputAPIAsHuman_NilAPI(t *testing.T) {
	err := outputAPIAsHuman(nil, "", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API data is nil")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_NoCustomDomain_NoUpstream covers L687 CustomDomain==""
// and L690 UpstreamURL=="" (both false branches).
func TestOutputAPIAsHuman_NoCustomDomain_NoUpstream(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API One", ListenPath: "/a/", DefaultVersion: "v1",
		CustomDomain: "", UpstreamURL: "",
	}
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	rErr, wErr, _ := os.Pipe()
	rOut, wOut, _ := os.Pipe()
	os.Stderr = wErr
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	stderrBytes, _ := io.ReadAll(rErr)
	_, _ = io.ReadAll(rOut)
	require.NoError(t, err)
	assert.NotContains(t, string(stderrBytes), "Custom Domain:")
	assert.NotContains(t, string(stderrBytes), "Upstream URL:")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_WithCustomDomainAndUpstream covers L687/L690 true branches.
func TestOutputAPIAsHuman_WithCustomDomainAndUpstream(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		CustomDomain: "api.example.com", UpstreamURL: "http://upstream:8080",
		OAS: mockOASAPIResponse(),
	}
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	rErr, wErr, _ := os.Pipe()
	rOut, wOut, _ := os.Pipe()
	os.Stderr = wErr
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	stderrBytes, _ := io.ReadAll(rErr)
	_, _ = io.ReadAll(rOut)
	require.NoError(t, err)
	assert.Contains(t, string(stderrBytes), "Custom Domain:")
	assert.Contains(t, string(stderrBytes), "api.example.com")
	assert.Contains(t, string(stderrBytes), "Upstream URL:")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_VersionMatchesDefault covers L703 versionName==DefaultVersion=T.
func TestOutputAPIAsHuman_VersionMatchesDefault(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS: mockOASAPIResponse(),
		VersionData: map[string]*types.APIVersion{
			"v1": {Name: "v1"},
			"v2": {Name: "v2"},
		},
	}
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	rErr, wErr, _ := os.Pipe()
	_, wOut, _ := os.Pipe()
	os.Stderr = wErr
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	stderrBytes, _ := io.ReadAll(rErr)
	require.NoError(t, err)
	// "(default)" marker should appear next to v1
	assert.Contains(t, string(stderrBytes), "(default)")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_RequestedVersion_Existing covers L717 requestedVersion!=""
// AND L719 versionData exists path.
func TestOutputAPIAsHuman_RequestedVersion_Existing(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS: mockOASAPIResponse(),
		VersionData: map[string]*types.APIVersion{
			"v2": {Name: "v2", OAS: map[string]interface{}{
				"openapi": "3.0.3",
				"info":    map[string]interface{}{"title": "T", "version": "2"},
			}},
		},
	}
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	rErr, wErr, _ := os.Pipe()
	rOut, wOut, _ := os.Pipe()
	os.Stderr = wErr
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "v2", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	stderrBytes, _ := io.ReadAll(rErr)
	_, _ = io.ReadAll(rOut)
	require.NoError(t, err)
	// versionToShow should be "v2" not "main", so the header should mention v2.
	assert.Contains(t, string(stderrBytes), "version: v2")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_RequestedVersion_FallbackToMain covers L719 versionData
// not exists (or OAS nil) path with a non-OAS-only call so the "Warning" branch
// emits to stderr (L726 !oasOnly=T).
func TestOutputAPIAsHuman_RequestedVersion_FallbackToMain(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS:         mockOASAPIResponse(),
		VersionData: map[string]*types.APIVersion{},
	}
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	rErr, wErr, _ := os.Pipe()
	rOut, wOut, _ := os.Pipe()
	os.Stderr = wErr
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "nonexistent", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	stderrBytes, _ := io.ReadAll(rErr)
	_, _ = io.ReadAll(rOut)
	require.NoError(t, err)
	assert.Contains(t, string(stderrBytes), "Warning")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_OASOnlyMode covers L678 !oasOnly=F AND L738 oasOnly=T paths.
func TestOutputAPIAsHuman_OASOnlyMode(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS: mockOASAPIResponse(),
	}
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	rErr, wErr, _ := os.Pipe()
	rOut, wOut, _ := os.Pipe()
	os.Stderr = wErr
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "", true)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	stderrBytes, _ := io.ReadAll(rErr)
	stdoutBytes, _ := io.ReadAll(rOut)
	require.NoError(t, err)
	// No summary on stderr; x-tyk-api-gateway stripped from YAML output.
	assert.NotContains(t, string(stderrBytes), "API Summary:")
	assert.NotContains(t, string(stdoutBytes), "x-tyk-api-gateway")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_RequestedVersion_FallbackToMain_OASOnly covers the
// L726 !oasOnly=F branch (oasOnly=T with a fallback to main).
func TestOutputAPIAsHuman_RequestedVersion_FallbackToMain_OASOnly(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS:         mockOASAPIResponse(),
		VersionData: map[string]*types.APIVersion{},
	}
	oldStderr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "missing", true /*oasOnly*/)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_RequestedVersion_NoVersionDataNoMain drives L722
// api.OAS != nil = F branch (requested version not in VersionData AND main
// OAS is nil).
func TestOutputAPIAsHuman_RequestedVersion_NoVersionDataNoMain(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS:         nil,
		VersionData: map[string]*types.APIVersion{},
	}
	oldStderr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "missing", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_RequestedVersion_DataOASNil drives L719 second-half
// branch where exists=T but VersionData[req].OAS==nil.
func TestOutputAPIAsHuman_RequestedVersion_DataOASNil(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS: mockOASAPIResponse(),
		VersionData: map[string]*types.APIVersion{
			"v3": {Name: "v3", OAS: nil},
		},
	}
	oldStderr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "v3", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_NilOASData_OASOnly covers L764 !oasOnly=F branch.
func TestOutputAPIAsHuman_NilOASData_OASOnly(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS: nil,
	}
	oldStderr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "", true /*oasOnly*/)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_RequestedVersionWithVersionDataOASNil covers L719
// exists && versionData.OAS != nil, where exists=T but OAS=nil (skipped is the
// short-circuit; we need T => T case).
func TestOutputAPIAsHuman_RequestedVersionWithOAS(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS: mockOASAPIResponse(),
		VersionData: map[string]*types.APIVersion{
			"v3": {Name: "v3", OAS: map[string]interface{}{
				"openapi": "3.0.3",
				"info":    map[string]interface{}{"title": "v3"},
			}},
		},
	}
	oldStderr := os.Stderr
	_, wErr, _ := os.Pipe()
	os.Stderr = wErr
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "v3", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// reqproof:req REQ-API-002
// TestOutputAPIAsHuman_NilOASData covers L763 oasData==nil path with !oasOnly=T.
func TestOutputAPIAsHuman_NilOASData(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", DefaultVersion: "v1",
		OAS: nil,
	}
	oldStderr := os.Stderr
	oldStdout := os.Stdout
	rErr, wErr, _ := os.Pipe()
	_, wOut, _ := os.Pipe()
	os.Stderr = wErr
	os.Stdout = wOut
	err := outputAPIAsHuman(api, "", false)
	wErr.Close()
	wOut.Close()
	os.Stderr = oldStderr
	os.Stdout = oldStdout
	stderrBytes, _ := io.ReadAll(rErr)
	require.NoError(t, err)
	assert.Contains(t, string(stderrBytes), "No OAS document available")
}

// reqproof:req REQ-API-002
// TestOutputAPIAsJSON_NoOAS covers L653 (oasOnly && api.OAS != nil)=F branch
// (oasOnly=T but OAS is nil so api.OAS != nil is false, so we encode the full
// api).
func TestOutputAPIAsJSON_OASOnlyButNilOAS(t *testing.T) {
	api := &types.OASAPI{
		ID: "a1", Name: "API", ListenPath: "/a/", OAS: nil,
	}
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputAPIAsJSON(api, true)
	wOut.Close()
	os.Stdout = oldStdout
	out, _ := io.ReadAll(rOut)
	require.NoError(t, err)
	// Output should be the full API (oasOnly skipped because OAS is nil), so
	// the top-level "id" field is present.
	assert.Contains(t, string(out), "\"id\": \"a1\"")
}

// ---------------------------------------------------------------------------
// MC/DC coverage for output*AsHuman / output*AsJSON variants with/without
// CustomDomain and UpstreamURL.
// ---------------------------------------------------------------------------

// reqproof:req REQ-API-003
func TestOutputCreatedAPIAsHuman_AllOptionalsSet(t *testing.T) {
	api := &types.OASAPI{
		ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1",
		CustomDomain: "api.example.com", UpstreamURL: "http://up:8080",
	}
	require.NoError(t, outputCreatedAPIAsHuman(api, "v1"))
}

// reqproof:req REQ-API-003
func TestOutputCreatedAPIAsHuman_NoOptionals(t *testing.T) {
	api := &types.OASAPI{
		ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1",
		CustomDomain: "", UpstreamURL: "",
	}
	require.NoError(t, outputCreatedAPIAsHuman(api, "v1"))
}

// reqproof:req REQ-API-003
func TestOutputCreatedAPIAsJSON_AllOptionalsSet(t *testing.T) {
	api := &types.OASAPI{
		ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1",
		CustomDomain: "api.example.com", UpstreamURL: "http://up:8080",
	}
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputCreatedAPIAsJSON(api, "v1")
	wOut.Close()
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// reqproof:req REQ-API-003
func TestOutputCreatedAPIAsJSON_NoOptionals(t *testing.T) {
	api := &types.OASAPI{ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1"}
	oldStdout := os.Stdout
	_, wOut, _ := os.Pipe()
	os.Stdout = wOut
	err := outputCreatedAPIAsJSON(api, "v1")
	wOut.Close()
	os.Stdout = oldStdout
	require.NoError(t, err)
}

// reqproof:req REQ-API-004
func TestOutputImportedAPIAsHuman_NoOptionals(t *testing.T) {
	api := &types.OASAPI{ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1"}
	require.NoError(t, outputImportedAPIAsHuman(api, "v1"))
}

// reqproof:req REQ-API-004
func TestOutputImportedAPIAsHuman_AllOptionals(t *testing.T) {
	api := &types.OASAPI{
		ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1",
		CustomDomain: "api.example.com", UpstreamURL: "http://up:8080",
	}
	require.NoError(t, outputImportedAPIAsHuman(api, "v1"))
}

// reqproof:req REQ-API-006
func TestOutputUpdatedAPIAsHuman_NoOptionals(t *testing.T) {
	api := &types.OASAPI{ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1"}
	require.NoError(t, outputUpdatedAPIAsHuman(api, "v1"))
}

// reqproof:req REQ-API-006
func TestOutputUpdatedAPIAsHuman_AllOptionals(t *testing.T) {
	api := &types.OASAPI{
		ID: "a", Name: "A", ListenPath: "/a/", DefaultVersion: "v1",
		CustomDomain: "api.example.com", UpstreamURL: "http://up:8080",
	}
	require.NoError(t, outputUpdatedAPIAsHuman(api, "v1"))
}

// reqproof:req REQ-API-007
func TestOutputDeletedAPIAsHuman_NoName(t *testing.T) {
	require.NoError(t, outputDeletedAPIAsHuman("api-id", ""))
}

// reqproof:req REQ-API-007
func TestOutputDeletedAPIAsHuman_WithName(t *testing.T) {
	require.NoError(t, outputDeletedAPIAsHuman("api-id", "API Name"))
}

// reqproof:req REQ-API-002
// TestRunAPIGet_NewClientFails covers L620 err!=nil from client.NewClient.
func TestRunAPIGet_NewClientFails(t *testing.T) {
	cmd := NewAPIGetCommand()
	cmd.SetContext(withConfig(context.Background(), brokenConfig()))
	cmd.SetArgs([]string{"x"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
}

// reqproof:req REQ-API-002
// TestRunAPIGet_ConfigNil covers L614 config==nil=T branch.
func TestRunAPIGet_ConfigNil(t *testing.T) {
	cmd := NewAPIGetCommand()
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"some-id"})
	err := cmd.RunE(cmd, []string{"some-id"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not found")
}

// reqproof:req REQ-API-002
// TestRunAPIGet_404OnlyMessage covers L632 substring branch where "404"
// matches but "not found" does not. Use a JSON body with a message that
// contains "404" alone.
func TestRunAPIGet_404OnlyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // 400 so status text won't contaminate
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 400, "message": "error 404 occurred",
		})
	}))
	defer server.Close()

	cmd := NewAPIGetCommand()
	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
	}}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputHuman))
	cmd.SetArgs([]string{"missing"})
	err := cmd.RunE(cmd, []string{"missing"})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 3, exitErr.Code)
}

// reqproof:req REQ-API-002
// TestRunAPIGet_GenericError covers L632 substring branches both F (the wrap
// fallback case where err.Error() contains neither "404" nor "not found").
func TestRunAPIGet_GenericError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 400, "message": "invalid id format",
		})
	}))
	defer server.Close()

	cmd := NewAPIGetCommand()
	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
	}}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputHuman))
	cmd.SetArgs([]string{"x"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
	// Should fall through to wrap (not ExitError code 3) since neither "404"
	// nor "not found" is in the message.
	if exitErr, ok := err.(*ExitError); ok {
		assert.NotEqual(t, 3, exitErr.Code)
	}
}

// reqproof:req REQ-API-002
// TestRunAPIGet_NotFoundOnlyText covers L632 substring branch where the error
// matches "not found" but not "404".
func TestRunAPIGet_NotFoundTextOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 400 with a not found phrasing so the substring branch fires.
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "message": "api was not found in the system"})
	}))
	defer server.Close()

	cmd := NewAPIGetCommand()
	cfg := &types.Config{DefaultEnvironment: "test", Environments: map[string]*types.Environment{
		"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
	}}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetContext(withOutputFormat(cmd.Context(), types.OutputHuman))
	cmd.SetArgs([]string{"missing"})
	err := cmd.RunE(cmd, []string{"missing"})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, 3, exitErr.Code)
}

// reqproof:req REQ-API-022
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