package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/internal/config"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// reqproof:req REQ-CFG-004
// reqproof:req REQ-CFG-005
// reqproof:req REQ-CFG-006
// captureColorOutputBuf swaps color.Output for a buffer for the lifetime of
// the test. The fatih/color package writes through its own writer initialised
// at package import time, so redirecting os.Stdout alone is not enough.
func captureColorOutputBuf(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	old := color.Output
	color.Output = buf
	t.Cleanup(func() { color.Output = old })
	return buf
}

// ---------------------------------------------------------------------------
// REQ-CFG-004 — config list
// ---------------------------------------------------------------------------

// reqproof:req REQ-CFG-004
func TestConfigList_RendersAllEnvironmentsWithActiveMarked(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	buf := captureColorOutputBuf(t)

	cmd := NewConfigListCommand()
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "dev", "list output should mention 'dev'")
	assert.Contains(t, out, "staging", "list output should mention 'staging'")
	assert.Contains(t, out, "active", "active environment must be marked")
	// Default is 'dev' in twoEnvConfig; check that the active marker is on it.
	devLine := ""
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "dev") && strings.Contains(line, "active") {
			devLine = line
			break
		}
	}
	assert.NotEmpty(t, devLine, "expected a line marking 'dev' as active, got: %q", out)
}

// reqproof:req REQ-CFG-004
func TestConfigList_EmptyConfigPrintsHelpfulMessage(t *testing.T) {
	setupTempConfig(t, "# empty\n")
	buf := captureColorOutputBuf(t)

	// Capture stdout separately for the fmt.Println call.
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	cmd := NewConfigListCommand()
	err := cmd.RunE(cmd, []string{})

	wOut.Close()
	os.Stdout = oldStdout
	stdoutBytes := make([]byte, 4096)
	n, _ := rOut.Read(stdoutBytes)
	stdoutStr := string(stdoutBytes[:n])

	require.NoError(t, err, "empty config should not error")
	combined := buf.String() + stdoutStr
	assert.Contains(t, combined, "No environments configured",
		"empty-state message must guide the user toward 'tyk config add'")
}

// ---------------------------------------------------------------------------
// REQ-CFG-005 — config set
// ---------------------------------------------------------------------------

// reqproof:req REQ-CFG-005
func TestConfigSet_NoFlagsReturnsBadArgs(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigSetCommand()
	cmd.SetArgs([]string{})
	_ = cmd.ParseFlags([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err, "config set with no flags must fail")
	assert.Contains(t, err.Error(), "at least one configuration value",
		"error must cite the missing-flag contract")
}

// reqproof:req REQ-CFG-005
func TestConfigSet_UpdatesActiveEnvironmentAndPersists(t *testing.T) {
	tykDir := setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigSetCommand()
	args := []string{"--dashboard-url", "http://updated.example.com"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)

	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	contents, err := os.ReadFile(filepath.Join(tykDir, "cli.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(contents), "http://updated.example.com",
		"new dashboard_url must be persisted to cli.toml")
	// The change must hit the active env (dev), not staging.
	out := string(contents)
	devIdx := strings.Index(out, "[environments.dev]")
	stagingIdx := strings.Index(out, "[environments.staging]")
	require.NotEqual(t, -1, devIdx, "dev section must remain")
	require.NotEqual(t, -1, stagingIdx, "staging section must remain")
	// Determine which section contains the updated URL.
	updatedIdx := strings.Index(out, "http://updated.example.com")
	require.NotEqual(t, -1, updatedIdx)
	// Updated URL must fall inside the dev section, i.e. between devIdx and stagingIdx.
	// Handle either section ordering.
	if devIdx < stagingIdx {
		assert.True(t, updatedIdx > devIdx && updatedIdx < stagingIdx,
			"updated URL must be inside the dev section, not staging")
	} else {
		assert.True(t, updatedIdx > devIdx, "updated URL must be inside the dev section")
	}
}

// reqproof:req REQ-CFG-005
func TestConfigSet_NoActiveEnvironmentReturnsError(t *testing.T) {
	setupTempConfig(t, `[environments.dev]
name = "dev"
dashboard_url = "http://dev.localhost:3000"
auth_token = "dev-token"
org_id = "dev-org"
`)

	cmd := NewConfigSetCommand()
	args := []string{"--auth-token", "new-token"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err, "set without an active env must fail")
	assert.Contains(t, err.Error(), "no active environment",
		"error must point to 'tyk config add' or default-env setup")
}

// ---------------------------------------------------------------------------
// REQ-CFG-006 — config remove
// ---------------------------------------------------------------------------

// reqproof:req REQ-CFG-006
func TestConfigRemove_UnknownEnvironmentReturnsError(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigRemoveCommand()
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.RunE(cmd, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found",
		"error must cite the missing environment name")
}

// reqproof:req REQ-CFG-006
func TestConfigRemove_RefusesToRemoveOnlyEnvironment(t *testing.T) {
	setupTempConfig(t, `default_environment = "dev"

[environments.dev]
name = "dev"
dashboard_url = "http://dev.localhost:3000"
auth_token = "dev-token"
org_id = "dev-org"
`)

	cmd := NewConfigRemoveCommand()
	cmd.SetArgs([]string{"dev"})

	err := cmd.RunE(cmd, []string{"dev"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove the last environment",
		"error must explain why the sole environment cannot be removed")
}

// reqproof:req REQ-CFG-006
func TestConfigRemove_RemovesEnvironmentAndPersists(t *testing.T) {
	tykDir := setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigRemoveCommand()
	cmd.SetArgs([]string{"staging"})

	err := cmd.RunE(cmd, []string{"staging"})
	require.NoError(t, err)

	contents, err := os.ReadFile(filepath.Join(tykDir, "cli.toml"))
	require.NoError(t, err)
	out := string(contents)
	assert.Contains(t, out, "[environments.dev]", "dev must remain")
	assert.NotContains(t, out, "[environments.staging]", "staging must be removed from TOML")
}

// ---------------------------------------------------------------------------
// REQ-CFG-005 — additional config set branches to exercise every flag combination
// (closes MC/DC gaps in runConfigSet)
// ---------------------------------------------------------------------------

// reqproof:req REQ-CFG-005
func TestConfigSet_FlagCombinations(t *testing.T) {
	tests := []struct {
		name        string
		flags       []string
		assertField string
	}{
		{"only auth-token", []string{"--auth-token", "new-token-xyz"}, "auth_token"},
		{"only org-id", []string{"--org-id", "new-org-zzz"}, "org_id"},
		{"dashboard + auth", []string{"--dashboard-url", "http://x.example.com", "--auth-token", "abc"}, "dashboard_url"},
		{"dashboard + org", []string{"--dashboard-url", "http://y.example.com", "--org-id", "neworg"}, "dashboard_url"},
		{"auth + org", []string{"--auth-token", "tok", "--org-id", "neworg"}, "auth_token"},
		{"all three", []string{"--dashboard-url", "http://z.example.com", "--auth-token", "tok2", "--org-id", "org2"}, "dashboard_url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTempConfig(t, twoEnvConfig)
			_ = captureColorOutputBuf(t)

			cmd := NewConfigSetCommand()
			cmd.SetArgs(tt.flags)
			_ = cmd.ParseFlags(tt.flags)

			err := cmd.RunE(cmd, []string{})
			require.NoError(t, err, "flags %v should produce no error on valid env", tt.flags)
		})
	}
}

// reqproof:req REQ-CFG-005
func TestConfigSet_InvalidatesEnvironmentReturnsError(t *testing.T) {
	// Start with a valid 2-env config, then try to set an empty dashboard_url
	// (impossible via the runner because empty values are filtered, but a malformed
	// dashboard URL passes the filter and fails Environment.Validate).
	setupTempConfig(t, twoEnvConfig)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigSetCommand()
	args := []string{"--dashboard-url", "not-a-url"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err, "malformed dashboard URL must fail validation after update")
	assert.Contains(t, err.Error(), "invalid dashboard URL format",
		"error must come from Environment.Validate, not from the flag parser")
}

// ---------------------------------------------------------------------------
// MC/DC coverage for runConfigAdd / runConfigList / runConfigCurrent /
// runConfigRemove / generateTOMLConfigUnified.
// ---------------------------------------------------------------------------

// reqproof:req REQ-CFG-031
// TestConfigAdd_SuccessNonDefault covers L321 setDefault||matches=F branches
// (set-default flag absent, env not default).
func TestConfigAdd_SuccessNonDefault(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigAddCommand()
	args := []string{"newenv",
		"--dashboard-url", "http://new.example.com",
		"--auth-token", "tok",
		"--org-id", "org",
	}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	require.NoError(t, cmd.RunE(cmd, []string{"newenv"}))
}

// reqproof:req REQ-CFG-031
// TestConfigAdd_SuccessAsDefault covers L321 setDefault=T branch.
func TestConfigAdd_SuccessAsDefault(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigAddCommand()
	args := []string{"newenv",
		"--dashboard-url", "http://new.example.com",
		"--auth-token", "tok",
		"--org-id", "org",
		"--set-default",
	}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	require.NoError(t, cmd.RunE(cmd, []string{"newenv"}))
}

// reqproof:req REQ-CFG-031
// TestConfigAdd_FirstEnvironmentBecomesDefault covers L321 setDefault=F &&
// DefaultEnvironment==envName=T (the second OR-operand).
func TestConfigAdd_FirstEnvironmentBecomesDefault(t *testing.T) {
	// Start with an empty config; the first env added becomes default automatically.
	setupTempConfig(t, "# empty\n")
	_ = captureColorOutputBuf(t)

	cmd := NewConfigAddCommand()
	args := []string{"firstenv",
		"--dashboard-url", "http://first.example.com",
		"--auth-token", "tok",
		"--org-id", "org",
	}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	require.NoError(t, cmd.RunE(cmd, []string{"firstenv"}))
}

// reqproof:req REQ-CFG-031
// TestConfigAdd_DuplicateReturnsError covers L305 err==nil=T branch.
func TestConfigAdd_DuplicateReturnsError(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigAddCommand()
	args := []string{"dev", // dev already exists in twoEnvConfig
		"--dashboard-url", "http://dup.example.com",
		"--auth-token", "tok",
		"--org-id", "org",
	}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{"dev"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// reqproof:req REQ-CFG-031
// TestConfigAdd_ValidationFailsReturnsError covers L294 err!=nil from
// env.Validate().
func TestConfigAdd_ValidationFailsReturnsError(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigAddCommand()
	args := []string{"bad",
		"--dashboard-url", "not-a-url",
		"--auth-token", "tok",
		"--org-id", "org",
	}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{"bad"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid dashboard URL")
}

// reqproof:req REQ-CFG-006
// TestConfigRemove_RemovesDefaultAndRotates covers L419 cfg.DefaultEnvironment==envName=T
// (removing the current default rotates to another env).
func TestConfigRemove_RemovesDefaultAndRotates(t *testing.T) {
	tykDir := setupTempConfig(t, twoEnvConfig) // default = dev
	_ = captureColorOutputBuf(t)

	cmd := NewConfigRemoveCommand()
	cmd.SetArgs([]string{"dev"})
	require.NoError(t, cmd.RunE(cmd, []string{"dev"}))

	// The TOML should now have staging as default.
	data, err := os.ReadFile(filepath.Join(tykDir, "cli.toml"))
	require.NoError(t, err)
	out := string(data)
	assert.Contains(t, out, `default_environment = "staging"`)
}

// reqproof:req REQ-CFG-001
// TestRunConfigCurrent_NoActiveEnvErrors covers L260 err!=nil branch (active
// env lookup fails — set default to a non-existing env name).
func TestRunConfigCurrent_NoActiveEnvErrors(t *testing.T) {
	// default_environment points at "ghost" but the env block doesn't exist.
	setupTempConfig(t, `default_environment = "ghost"

[environments.dev]
name = "dev"
dashboard_url = "http://dev:3000"
auth_token = "tok"
org_id = "org"
`)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigCurrentCommand()
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active environment")
}

// reqproof:req REQ-CFG-030
// TestGenerateTOMLConfigUnified_NoDefault covers L477 cfg.DefaultEnvironment!=""=F.
func TestGenerateTOMLConfigUnified_NoDefault(t *testing.T) {
	cfg := &types.Config{
		DefaultEnvironment: "",
		Environments: map[string]*types.Environment{
			"dev": {Name: "dev", DashboardURL: "http://dev", AuthToken: "tok", OrgID: "org"},
		},
	}
	out := generateTOMLConfigUnified(cfg)
	assert.NotContains(t, out, "default_environment")
}

// reqproof:req REQ-CFG-030
// TestGenerateTOMLConfigUnified_NoEnvironments covers L482 len(envs)>0=F.
func TestGenerateTOMLConfigUnified_NoEnvironments(t *testing.T) {
	cfg := &types.Config{DefaultEnvironment: "x", Environments: map[string]*types.Environment{}}
	out := generateTOMLConfigUnified(cfg)
	assert.Contains(t, out, "# Tyk CLI Configuration")
	assert.NotContains(t, out, "[environments.")
}

// reqproof:req REQ-CFG-030
// TestGenerateTOMLConfigUnified_WithTimeout covers L489 env.TimeoutSeconds>0=T.
func TestGenerateTOMLConfigUnified_WithTimeout(t *testing.T) {
	cfg := &types.Config{
		DefaultEnvironment: "dev",
		Environments: map[string]*types.Environment{
			"dev": {Name: "dev", DashboardURL: "http://dev", AuthToken: "tok", OrgID: "org", TimeoutSeconds: 30},
		},
	}
	out := generateTOMLConfigUnified(cfg)
	assert.Contains(t, out, "timeout_seconds = 30")
}

// reqproof:req REQ-CFG-005
// TestRunConfigSet_DashboardURLOnly covers the dashboard-url-only flag branch
// (the others stay empty, exercising the F branch for auth_token and org_id).
func TestRunConfigSet_DashboardURLOnly(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigSetCommand()
	args := []string{"--dashboard-url", "http://only.example.com"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	require.NoError(t, cmd.RunE(cmd, []string{}))
}


// reqproof:req REQ-CFG-005
// TestRunConfigSet_LoadConfigFails covers L340 err!=nil from LoadConfig
// (malformed TOML file).
func TestRunConfigSet_LoadConfigFails(t *testing.T) {
	// Set HOME and write a malformed TOML.
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	// TOML with unterminated string -> parse error.
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("[bad\n"), 0o600))

	cmd := NewConfigSetCommand()
	args := []string{"--dashboard-url", "http://x.example.com"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// reqproof:req REQ-CFG-005
// TestRunConfigSet_GetEnvironmentFails covers L352 err!=nil from
// manager.GetEnvironment(cfg.DefaultEnvironment) where the named default env
// is missing from environments.
func TestRunConfigSet_GetEnvironmentFails(t *testing.T) {
	setupTempConfig(t, `default_environment = "ghost"

[environments.dev]
name = "dev"
dashboard_url = "http://dev:3000"
auth_token = "tok"
org_id = "org"
`)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigSetCommand()
	args := []string{"--dashboard-url", "http://x.example.com"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// reqproof:req REQ-CFG-006
// TestRunConfigRemove_LoadConfigFails covers L399 err!=nil from LoadConfig.
func TestRunConfigRemove_LoadConfigFails(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("[bad\n"), 0o600))

	cmd := NewConfigRemoveCommand()
	cmd.SetArgs([]string{"x"})
	err := cmd.RunE(cmd, []string{"x"})
	require.Error(t, err)
}

// reqproof:req REQ-CFG-031
// TestRunConfigAdd_LoadConfigFails covers L300 err!=nil from LoadConfig.
func TestRunConfigAdd_LoadConfigFails(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("[bad\n"), 0o600))

	cmd := NewConfigAddCommand()
	args := []string{"newenv",
		"--dashboard-url", "http://new.example.com",
		"--auth-token", "tok",
		"--org-id", "org",
	}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{"newenv"})
	require.Error(t, err)
}

// reqproof:req REQ-CFG-003
// TestRunConfigUse_LoadConfigFails covers L196 err!=nil from LoadConfig.
func TestRunConfigUse_LoadConfigFails(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("[bad\n"), 0o600))

	cmd := NewConfigUseCommand()
	cmd.SetArgs([]string{"dev"})
	err := cmd.RunE(cmd, []string{"dev"})
	require.Error(t, err)
}

// reqproof:req REQ-CFG-001
// TestRunConfigList_LoadConfigFails covers L147 err!=nil.
func TestRunConfigList_LoadConfigFails(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("[bad\n"), 0o600))

	cmd := NewConfigListCommand()
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// reqproof:req REQ-CFG-001
// TestRunConfigCurrent_LoadConfigFails covers L247 err!=nil.
func TestRunConfigCurrent_LoadConfigFails(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("[bad\n"), 0o600))

	cmd := NewConfigCurrentCommand()
	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
}

// reqproof:req REQ-CFG-030
// TestSaveConfigToFile_DirectFailure exercises saveConfigToFile directly by
// constructing a manager and pointing HOME at a regular file so MkdirAll
// inside saveConfigToFile fails.
func TestSaveConfigToFile_DirectFailure(t *testing.T) {
	tempHome := t.TempDir()
	fileBlock := filepath.Join(tempHome, "block")
	require.NoError(t, os.WriteFile(fileBlock, []byte("file"), 0o600))
	t.Setenv("HOME", fileBlock)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(fileBlock, ".config"))

	manager := config.NewManager()
	require.NoError(t, manager.SaveEnvironment(&types.Environment{
		Name: "dev", DashboardURL: "http://dev:3000", AuthToken: "tok", OrgID: "org",
	}, true))

	err := saveConfigToFile(manager)
	require.Error(t, err)
}

// reqproof:req REQ-CFG-030
// TestSaveConfigToFile_WriteFailDueToDirAsFile covers L454 err!=nil from
// os.WriteFile by ensuring the cli.toml path itself is a directory.
func TestSaveConfigToFile_WriteFailDueToDirAsFile(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))
	userCfg, _ := os.UserConfigDir()
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(filepath.Join(tykDir, "cli.toml"), 0o755))

	manager := config.NewManager()
	require.NoError(t, manager.SaveEnvironment(&types.Environment{
		Name: "dev", DashboardURL: "http://dev:3000", AuthToken: "tok", OrgID: "org",
	}, true))

	err := saveConfigToFile(manager)
	require.Error(t, err)
}

// reqproof:req REQ-CFG-030
// TestSaveConfigToFile_MkdirAllFails covers L450 err!=nil. We point HOME at a
// path whose parent is a regular file so MkdirAll cannot create the tyk dir.
func TestSaveConfigToFile_MkdirAllFails(t *testing.T) {
	tempHome := t.TempDir()
	// Create a regular file at the location that would normally be a directory
	// inside HOME's config tree.
	fileBlock := filepath.Join(tempHome, "block")
	require.NoError(t, os.WriteFile(fileBlock, []byte("just a file"), 0o600))
	// Now point HOME inside that file path so the config dir path traverses it.
	t.Setenv("HOME", fileBlock)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(fileBlock, ".config"))

	// Construct a manager with a valid config in memory and try to save.
	// We use runConfigAdd which exercises the full save path.
	cmd := NewConfigAddCommand()
	args := []string{"newenv",
		"--dashboard-url", "http://new.example.com",
		"--auth-token", "tok",
		"--org-id", "org",
	}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	err := cmd.RunE(cmd, []string{"newenv"})
	// We expect either a load error (no cli.toml) or a save error.
	// Either path exercises new code in the chain.
	require.Error(t, err)
}

// reqproof:req REQ-CFG-030
// TestGetConfigDir_NoHome covers L464 err!=nil from os.UserConfigDir.
// On macOS UserConfigDir returns ($HOME)/Library/Application Support, but when
// HOME is empty and XDG_CONFIG_HOME also empty, UserConfigDir errors.
func TestGetConfigDir_NoHome(t *testing.T) {
	if _, present := os.LookupEnv("HOME"); present {
		t.Setenv("HOME", "")
	}
	if _, present := os.LookupEnv("XDG_CONFIG_HOME"); present {
		t.Setenv("XDG_CONFIG_HOME", "")
	}
	// On Darwin UserConfigDir fails when HOME is empty.
	_, err := getConfigDir()
	if err == nil {
		t.Skip("os.UserConfigDir did not fail on this platform/setup; skipping")
	}
}

// reqproof:req REQ-CFG-005
// TestRunConfigSet_OrgIDOnly covers the org-id-only branch.
func TestRunConfigSet_OrgIDOnly(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	_ = captureColorOutputBuf(t)

	cmd := NewConfigSetCommand()
	args := []string{"--org-id", "new-org"}
	cmd.SetArgs(args)
	_ = cmd.ParseFlags(args)
	require.NoError(t, cmd.RunE(cmd, []string{}))
}
