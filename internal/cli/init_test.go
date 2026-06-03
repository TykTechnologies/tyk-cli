package cli

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// ---------------------------------------------------------------------------
// MC/DC coverage for init.go helpers (askString / askYesNo / gatherEnvironmentInfo /
// saveEnvironment / testConnection / runQuickSetup).
// ---------------------------------------------------------------------------

// reqproof:req REQ-CFG-002
// scannerFor returns a bufio.Scanner reading from the supplied string.
func scannerFor(s string) *bufio.Scanner {
	return bufio.NewScanner(strings.NewReader(s))
}

// reqproof:req REQ-CFG-002
// TestAskString_PromptOnlyNoDefault drives L220 defaultValue!=""=F branch and
// L229 input==""=F branch (returns input).
func TestAskString_PromptOnlyNoDefault(t *testing.T) {
	got := askString(scannerFor("hello\n"), "Q", "")
	assert.Equal(t, "hello", got)
}

// reqproof:req REQ-CFG-002
// TestAskString_DefaultUsed drives L220 defaultValue!=""=T branch and L229
// input==""=T && defaultValue!=""=T branch (returns default).
func TestAskString_DefaultUsed(t *testing.T) {
	got := askString(scannerFor("\n"), "Q", "fallback")
	assert.Equal(t, "fallback", got)
}

// reqproof:req REQ-CFG-002
// TestAskString_InputOverridesDefault drives L229 input==""=F branch with a
// default present (so input wins).
func TestAskString_InputOverridesDefault(t *testing.T) {
	got := askString(scannerFor("override\n"), "Q", "default")
	assert.Equal(t, "override", got)
}

// reqproof:req REQ-CFG-002
// TestAskYesNo_Yes covers the y/yes input path.
func TestAskYesNo_Yes(t *testing.T) {
	assert.True(t, askYesNo(scannerFor("y\n"), "Continue?"))
	assert.True(t, askYesNo(scannerFor("yes\n"), "Continue?"))
}

// reqproof:req REQ-CFG-002
// TestAskYesNo_No covers the non-yes input path.
func TestAskYesNo_No(t *testing.T) {
	assert.False(t, askYesNo(scannerFor("n\n"), "Continue?"))
	assert.False(t, askYesNo(scannerFor("\n"), "Continue?"))
}

// reqproof:req REQ-CFG-002
// TestGatherEnvironmentInfo_AllFieldsSet covers L102/L117 isFirst=T branches,
// and L112/L124/L136 empty-check branches all =F (happy path).
func TestGatherEnvironmentInfo_AllFieldsSet(t *testing.T) {
	input := "http://dash.local:3000\ntok-123\norg-456\n"
	env, err := gatherEnvironmentInfo(scannerFor(input), "dev", true)
	require.NoError(t, err)
	assert.Equal(t, "http://dash.local:3000", env.DashboardURL)
	assert.Equal(t, "tok-123", env.AuthToken)
	assert.Equal(t, "org-456", env.OrgID)
}

// reqproof:req REQ-CFG-002
// TestGatherEnvironmentInfo_MissingDashboardURL covers L112 env.DashboardURL=="" =T.
func TestGatherEnvironmentInfo_MissingDashboardURL(t *testing.T) {
	input := "\n" // empty dashboard URL line
	_, err := gatherEnvironmentInfo(scannerFor(input), "dev", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dashboard URL is required")
}

// reqproof:req REQ-CFG-002
// TestGatherEnvironmentInfo_MissingAuthToken covers L124 env.AuthToken=="" =T.
func TestGatherEnvironmentInfo_MissingAuthToken(t *testing.T) {
	input := "http://dash.local:3000\n\n"
	_, err := gatherEnvironmentInfo(scannerFor(input), "dev", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth token is required")
}

// reqproof:req REQ-CFG-002
// TestGatherEnvironmentInfo_MissingOrgID covers L136 env.OrgID=="" =T.
func TestGatherEnvironmentInfo_MissingOrgID(t *testing.T) {
	input := "http://dash.local:3000\ntok\n\n"
	_, err := gatherEnvironmentInfo(scannerFor(input), "dev", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization ID is required")
}

// reqproof:req REQ-CFG-002
// TestGatherEnvironmentInfo_NotFirst covers L102/L117 isFirst=F branch
// (no example output, but the helper still fills the env).
func TestGatherEnvironmentInfo_NotFirst(t *testing.T) {
	input := "http://dash.local:3000\ntok\norg\n"
	env, err := gatherEnvironmentInfo(scannerFor(input), "staging", false)
	require.NoError(t, err)
	assert.Equal(t, "http://dash.local:3000", env.DashboardURL)
}

// reqproof:req REQ-CFG-002
// withHomeDir redirects HOME / XDG_CONFIG_HOME / USERPROFILE so config dir
// lookups land in a per-test temp directory.
func withHomeDir(t *testing.T) string {
	t.Helper()
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	return tempHome
}

// reqproof:req REQ-CFG-002
// TestSaveEnvironment_NewFile covers the happy path: no existing config,
// successful Save + write to file. Drives L175 err==nil=F branch (file does
// not exist).
func TestSaveEnvironment_NewFile(t *testing.T) {
	_ = withHomeDir(t)
	env := &types.Environment{
		Name: "dev", DashboardURL: "http://dash:3000", AuthToken: "tok", OrgID: "org",
	}
	require.NoError(t, saveEnvironment(env, true))

	// Verify file was written.
	userCfg, err := os.UserConfigDir()
	require.NoError(t, err)
	configFile := filepath.Join(userCfg, "tyk", "cli.toml")
	data, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "dashboard_url = \"http://dash:3000\"")
}

// reqproof:req REQ-CFG-002
// TestSaveEnvironment_ExistingFile covers L175 err==nil=T branch (file exists,
// LoadConfig is called).
func TestSaveEnvironment_ExistingFile(t *testing.T) {
	_ = withHomeDir(t)

	// Pre-create config dir and file with a valid TOML.
	userCfg, err := os.UserConfigDir()
	require.NoError(t, err)
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	existing := `default_environment = "old"

[environments.old]
name = "old"
dashboard_url = "http://old:3000"
auth_token = "tok"
org_id = "org"
`
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte(existing), 0o600))

	// Add a new env.
	env := &types.Environment{
		Name: "new", DashboardURL: "http://new:3000", AuthToken: "tok2", OrgID: "org2",
	}
	require.NoError(t, saveEnvironment(env, true))

	data, err := os.ReadFile(filepath.Join(tykDir, "cli.toml"))
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, "[environments.old]")
	assert.Contains(t, out, "[environments.new]")
}

// reqproof:req REQ-CFG-002
// TestSaveEnvironment_WriteFailDueToDirAsFile covers L194 err!=nil from
// os.WriteFile by ensuring the cli.toml path itself is a directory.
func TestSaveEnvironment_WriteFailDueToDirAsFile(t *testing.T) {
	_ = withHomeDir(t)

	// Create cli.toml as a directory (not a regular file).
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(filepath.Join(tykDir, "cli.toml"), 0o755))

	env := &types.Environment{
		Name: "dev", DashboardURL: "http://dash:3000", AuthToken: "tok", OrgID: "org",
	}
	err := saveEnvironment(env, true)
	require.Error(t, err)
}

// reqproof:req REQ-CFG-002
// TestSaveEnvironment_GetConfigDirFails covers L167 err!=nil branch by
// unsetting HOME so UserConfigDir errors.
func TestSaveEnvironment_GetConfigDirFails(t *testing.T) {
	if _, hasHome := os.LookupEnv("HOME"); hasHome {
		oldHome := os.Getenv("HOME")
		os.Unsetenv("HOME")
		defer os.Setenv("HOME", oldHome)
	}
	if _, has := os.LookupEnv("XDG_CONFIG_HOME"); has {
		oldX := os.Getenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", oldX)
	}

	env := &types.Environment{
		Name: "dev", DashboardURL: "http://dash:3000", AuthToken: "tok", OrgID: "org",
	}
	err := saveEnvironment(env, true)
	if err == nil {
		t.Skip("os.UserConfigDir did not fail on this platform; skipping")
	}
}

// reqproof:req REQ-CFG-002
// TestSaveEnvironment_MkdirAllFails covers L190 err!=nil from os.MkdirAll by
// forcing the config dir to traverse a regular file.
func TestSaveEnvironment_MkdirAllFails(t *testing.T) {
	tempHome := t.TempDir()
	fileBlock := filepath.Join(tempHome, "block")
	require.NoError(t, os.WriteFile(fileBlock, []byte("file"), 0o600))
	t.Setenv("HOME", fileBlock)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(fileBlock, ".config"))

	env := &types.Environment{
		Name: "dev", DashboardURL: "http://dash:3000", AuthToken: "tok", OrgID: "org",
	}
	err := saveEnvironment(env, true)
	require.Error(t, err)
}

// reqproof:req REQ-CFG-002
// TestSaveEnvironment_ExistingMalformedConfigFails covers the L176 err!=nil branch
// from LoadConfig (existing file but malformed).
func TestSaveEnvironment_ExistingMalformedConfigFails(t *testing.T) {
	_ = withHomeDir(t)

	userCfg, err := os.UserConfigDir()
	require.NoError(t, err)
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("[bad\n"), 0o600))

	env := &types.Environment{
		Name: "new", DashboardURL: "http://new:3000", AuthToken: "tok", OrgID: "org",
	}
	err = saveEnvironment(env, true)
	require.Error(t, err)
}

// reqproof:req REQ-CFG-002
// TestTestConnection_Reachable covers L153 err!=nil=F branch (success).
func TestTestConnection_Reachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	env := &types.Environment{
		Name: "test", DashboardURL: server.URL, AuthToken: "tok", OrgID: "org",
	}
	require.NoError(t, testConnection(env))
}

// reqproof:req REQ-CFG-002
// TestTestConnection_NewClientFails covers L153 err!=nil=T branch where
// client.NewClient itself fails (invalid env causes Validate to return error).
func TestTestConnection_NewClientFails(t *testing.T) {
	// Validate fails on empty dashboard URL.
	env := &types.Environment{Name: "test"}
	err := testConnection(env)
	require.Error(t, err)
}

// reqproof:req REQ-CFG-002
// TestTestConnection_Unreachable covers L153 err!=nil=T branch (failure).
func TestTestConnection_Unreachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	server.Close()

	env := &types.Environment{
		Name: "test", DashboardURL: server.URL, AuthToken: "tok", OrgID: "org",
	}
	require.Error(t, testConnection(env))
}

// reqproof:req REQ-CFG-002
// TestRunQuickSetup_SkipTest covers L75 !skipTest=F branch (skipTest=T).
func TestRunQuickSetup_SkipTest(t *testing.T) {
	_ = withHomeDir(t)
	input := "http://dash:3000\ntok\norg\n"
	err := runQuickSetup(scannerFor(input), true /* skipTest */)
	require.NoError(t, err)
}

// reqproof:req REQ-CFG-002
// TestRunQuickSetup_TestSucceeds covers L75 !skipTest=T && testConnection
// returns nil (the success branch which logs ✅).
func TestRunQuickSetup_TestSucceeds(t *testing.T) {
	_ = withHomeDir(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	input := server.URL + "\ntok\norg\n"
	err := runQuickSetup(scannerFor(input), false /* skipTest */)
	require.NoError(t, err)
}

// reqproof:req REQ-CFG-002
// TestRunQuickSetup_TestFails_Continue covers the failure-but-continue branch
// (testConnection fails, user answers "y").
func TestRunQuickSetup_TestFails_Continue(t *testing.T) {
	_ = withHomeDir(t)
	// Closed server to force connection failure.
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	failing.Close()

	input := failing.URL + "\ntok\norg\ny\n" // dash url, tok, org, then "y" to continue
	err := runQuickSetup(scannerFor(input), false)
	require.NoError(t, err)
}

// reqproof:req REQ-CFG-002
// TestRunQuickSetup_SaveEnvFails covers L86 err!=nil branch by setting HOME to
// a path whose parent is a regular file (MkdirAll fails inside saveEnvironment).
func TestRunQuickSetup_SaveEnvFails(t *testing.T) {
	tempHome := t.TempDir()
	fileBlock := filepath.Join(tempHome, "block")
	require.NoError(t, os.WriteFile(fileBlock, []byte("file"), 0o600))
	t.Setenv("HOME", fileBlock)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(fileBlock, ".config"))

	input := "http://dash:3000\ntok\norg\n"
	err := runQuickSetup(scannerFor(input), true /* skipTest */)
	require.Error(t, err)
}

// reqproof:req REQ-CFG-002
// TestRunQuickSetup_TestFails_Cancel covers the failure-and-cancel branch.
func TestRunQuickSetup_TestFails_Cancel(t *testing.T) {
	_ = withHomeDir(t)
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	failing.Close()

	input := failing.URL + "\ntok\norg\nn\n"
	err := runQuickSetup(scannerFor(input), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "setup cancelled")
}

// reqproof:req REQ-CFG-002
// TestRunQuickSetup_GatherFails covers L71 err!=nil=T branch (gather fails).
func TestRunQuickSetup_GatherFails(t *testing.T) {
	input := "\n" // empty dashboard URL line triggers gatherEnvironmentInfo error
	err := runQuickSetup(scannerFor(input), false)
	require.Error(t, err)
}
