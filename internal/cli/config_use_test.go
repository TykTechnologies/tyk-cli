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
)

// Verifies: SYS-REQ-043
// captureColorOutput swaps color.Output for a buffer for the duration of the
// test, returning the buffer. The fatih/color package writes through its own
// writer (initialized at init time), so redirecting os.Stdout alone is not
// enough to capture color output.
func captureColorOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	old := color.Output
	color.Output = buf
	t.Cleanup(func() { color.Output = old })
	return buf
}

// Verifies: SYS-REQ-043
// setupTempConfig redirects the user config dir to a temp directory for the
// duration of the test and writes the supplied TOML to cli.toml. Returns the
// config dir (the parent that contains cli.toml).
func setupTempConfig(t *testing.T, toml string) string {
	t.Helper()

	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	// Linux honors XDG_CONFIG_HOME; macOS uses $HOME/Library/Application Support.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempHome, ".config"))

	// Compute the platform-correct config dir under the redirected HOME.
	userCfg, err := os.UserConfigDir()
	require.NoError(t, err)
	tykDir := filepath.Join(userCfg, "tyk")
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte(toml), 0o600))

	return tykDir
}

// Verifies: SYS-REQ-043
const twoEnvConfig = `default_environment = "dev"

[environments.dev]
name = "dev"
dashboard_url = "http://dev.localhost:3000"
auth_token = "dev-token"
org_id = "dev-org"

[environments.staging]
name = "staging"
dashboard_url = "http://staging.example.com"
auth_token = "staging-token"
org_id = "staging-org"
`

// Verifies: SYS-REQ-043
func TestConfigUse_SwitchesActiveEnvironmentAndPersists(t *testing.T) {
	tykDir := setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigUseCommand()
	cmd.SetArgs([]string{"staging"})

	err := cmd.RunE(cmd, []string{"staging"})
	require.NoError(t, err)

	// Re-read the file via a fresh manager to confirm persistence.
	manager := config.NewManager()
	require.NoError(t, manager.LoadConfig())
	cfg := manager.GetConfig()
	assert.Equal(t, "staging", cfg.DefaultEnvironment, "default environment should be persisted as staging")

	// Sanity: the file on disk also reflects the switch.
	contents, err := os.ReadFile(filepath.Join(tykDir, "cli.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(contents), `default_environment = "staging"`)
}

// Verifies: SYS-REQ-043
// TestConfigUse_NoArgsInteractive covers L210 len(args)>0=F branch. Without
// args runConfigUse falls into selectEnvironmentInteractively which fails
// without a TTY; the failure path proves the F branch was taken.
func TestConfigUse_NoArgsInteractive(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigUseCommand()
	cmd.SetArgs([]string{})
	err := cmd.RunE(cmd, []string{})
	// We expect an error from the interactive prompt because the test harness
	// has no TTY; this proves the F branch of L210 was entered.
	require.Error(t, err)
}

// Verifies: SYS-REQ-043
func TestConfigUse_UnknownEnvironmentReturnsError(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)

	cmd := NewConfigUseCommand()
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.RunE(cmd, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "nonexistent",
		"error should mention the unknown environment name")
}

// Verifies: SYS-REQ-043
func TestConfigUse_NoEnvironmentsConfiguredReturnsError(t *testing.T) {
	setupTempConfig(t, "# empty\n")

	cmd := NewConfigUseCommand()
	cmd.SetArgs([]string{"anything"})

	err := cmd.RunE(cmd, []string{"anything"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no environments configured")
}

// Verifies: SYS-REQ-043
func TestConfigCurrent_ShowsActiveEnvironment(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	buf := captureColorOutput(t)

	cmd := NewConfigCurrentCommand()
	err := cmd.RunE(cmd, []string{})

	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "dev", "current env should mention 'dev'")
	assert.Contains(t, out, "http://dev.localhost:3000", "current env should mention dashboard_url")
	assert.Contains(t, out, "dev-org", "current env should mention org_id")
	// Token should be masked, not raw.
	assert.NotContains(t, out, "dev-token", "auth_token should be masked")
}


// Verifies: SYS-REQ-043
func TestConfigCurrent_NoDefaultEnvironmentReportsCleanly(t *testing.T) {
	setupTempConfig(t, `[environments.dev]
name = "dev"
dashboard_url = "http://dev.localhost:3000"
auth_token = "dev-token"
org_id = "dev-org"
`)
	buf := captureColorOutput(t)

	cmd := NewConfigCurrentCommand()
	err := cmd.RunE(cmd, []string{})

	require.NoError(t, err, "config current should not error when no default env is set")
	assert.Contains(t, buf.String(), "No default environment set")
}
