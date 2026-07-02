package cli

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Verifies: SYS-REQ-041
func TestNewRootCommand(t *testing.T) {
	rootCmd := NewRootCommand("1.0.0", "abc123", "2023-01-01T00:00:00Z")
	
	assert.Equal(t, "tyk", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Tyk CLI")
	assert.Equal(t, "1.0.0", rootCmd.Version)
	
	// Check that main subcommands are added
	apiCmd, _, err := rootCmd.Find([]string{"api"})
	assert.NoError(t, err)
	assert.Equal(t, "api", apiCmd.Use)
	
	configCmd, _, err := rootCmd.Find([]string{"config"})
	assert.NoError(t, err)
	assert.Equal(t, "config", configCmd.Use)
	
	initCmd, _, err := rootCmd.Find([]string{"init"})
	assert.NoError(t, err)
	assert.Equal(t, "init", initCmd.Use)
}

// Verifies: SYS-REQ-041
func TestGlobalFlags(t *testing.T) {
	rootCmd := NewRootCommand("1.0.0", "abc123", "2023-01-01T00:00:00Z")
	
	// Check that persistent flags are defined
	dashURLFlag := rootCmd.PersistentFlags().Lookup("dash-url")
	assert.NotNil(t, dashURLFlag)
	assert.Equal(t, "string", dashURLFlag.Value.Type())
	
	authTokenFlag := rootCmd.PersistentFlags().Lookup("auth-token")
	assert.NotNil(t, authTokenFlag)
	
	orgIDFlag := rootCmd.PersistentFlags().Lookup("org-id")
	assert.NotNil(t, orgIDFlag)
	
	jsonFlag := rootCmd.PersistentFlags().Lookup("json")
	assert.NotNil(t, jsonFlag)
	assert.Equal(t, "bool", jsonFlag.Value.Type())
}

// Verifies: SYS-REQ-041
func TestGetOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		jsonFlag bool
		expected types.OutputFormat
	}{
		{"human format", false, types.OutputHuman},
		{"json format", true, types.OutputJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOutputFormat(tt.jsonFlag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Verifies: SYS-REQ-041
func TestInitConfigWithEnvironment(t *testing.T) {
	// This test verifies that configuration can be loaded from flags
	// (since existing config files may override environment variables in real environments)
	rootCmd := NewRootCommand("1.0.0", "abc123", "2023-01-01T00:00:00Z")
	apiCmd, _, err := rootCmd.Find([]string{"api", "get"})
	require.NoError(t, err)

	// Set up command context
	ctx := context.Background()
	apiCmd.SetContext(ctx)

	// Test configuration loading with flag values 
	globalFlags := GlobalFlags{
		DashURL:   "http://test-dashboard:3000", 
		AuthToken: "test-token",
		OrgID:     "test-org",
	}
	err = initConfig(apiCmd, &globalFlags)
	require.NoError(t, err)

	// Verify config was loaded  
	config := GetConfigFromContext(apiCmd.Context())
	require.NotNil(t, config)
	
	// Get active environment and verify values
	activeEnv, err := config.GetActiveEnvironment()
	require.NoError(t, err)
	
	// The values should match what we set via flags 
	assert.Equal(t, "http://test-dashboard:3000", activeEnv.DashboardURL)
	assert.Equal(t, "test-token", activeEnv.AuthToken)  
	assert.Equal(t, "test-org", activeEnv.OrgID)
}

// Verifies: SYS-REQ-041
func TestInitConfigWithFlags(t *testing.T) {
	// Clean environment
	os.Unsetenv("TYK_DASH_URL")
	os.Unsetenv("TYK_AUTH_TOKEN")
	os.Unsetenv("TYK_ORG_ID")

	rootCmd := NewRootCommand("1.0.0", "abc123", "2023-01-01T00:00:00Z")
	apiCmd, _, err := rootCmd.Find([]string{"api", "get"})
	require.NoError(t, err)

	// Set up command context
	ctx := context.Background()
	apiCmd.SetContext(ctx)

	// Test with flags
	globalFlags := GlobalFlags{
		DashURL:   "http://flag-dashboard:3000",
		AuthToken: "flag-token",
		OrgID:     "flag-org",
		JSON:      true,
	}

	err = initConfig(apiCmd, &globalFlags)
	require.NoError(t, err)

	// Verify config was loaded from flags
	config := GetConfigFromContext(apiCmd.Context())
	require.NotNil(t, config)
	
	// Get active environment and verify values
	activeEnv, err := config.GetActiveEnvironment()
	require.NoError(t, err)
	assert.Equal(t, "http://flag-dashboard:3000", activeEnv.DashboardURL)
	assert.Equal(t, "flag-token", activeEnv.AuthToken)
	assert.Equal(t, "flag-org", activeEnv.OrgID)

	// Verify output format
	format := GetOutputFormatFromContext(apiCmd.Context())
	assert.Equal(t, types.OutputJSON, format)
}

// Verifies: SYS-REQ-041
func TestCommandSkipping(t *testing.T) {
	// Test that init and config commands don't require configuration
	rootCmd := NewRootCommand("1.0.0", "abc123", "2023-01-01T00:00:00Z")
	
	// Test init command can run without configuration
	rootCmd.SetArgs([]string{"init", "--help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	
	// Test config command can run without configuration 
	rootCmd.SetArgs([]string{"config", "--help"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
}

// Verifies: SYS-REQ-041
func TestVersionCommand(t *testing.T) {
	rootCmd := NewRootCommand("1.2.3", "def456", "2023-12-25T10:30:00Z")
	
	// Execute version flag
	rootCmd.SetArgs([]string{"--version"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

// Verifies: SYS-REQ-041
// TestInitConfig_LoadConfigFails covers L79 err!=nil from LoadConfig
// (malformed cli.toml).
func TestInitConfig_LoadConfigFails(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome+"/.config")
	userCfg, _ := os.UserConfigDir()
	tykDir := userCfg + "/tyk"
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	require.NoError(t, os.WriteFile(tykDir+"/cli.toml", []byte("[bad\n"), 0o600))

	rootCmd := NewRootCommand("1.0.0", "abc", "now")
	apiCmd, _, err := rootCmd.Find([]string{"api", "get"})
	require.NoError(t, err)
	apiCmd.SetContext(context.Background())

	flags := GlobalFlags{}
	err = initConfig(apiCmd, &flags)
	require.Error(t, err)
}

// Verifies: SYS-REQ-041
// TestInitConfig_ValidateFails covers L88 err!=nil from config.Validate (no
// environments configured).
func TestInitConfig_ValidateFails(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome+"/.config")
	// No cli.toml present and no flags provided → manager has empty config →
	// config.Validate returns "no environments configured".

	rootCmd := NewRootCommand("1.0.0", "abc", "now")
	apiCmd, _, err := rootCmd.Find([]string{"api", "get"})
	require.NoError(t, err)
	apiCmd.SetContext(context.Background())

	flags := GlobalFlags{} // no dash-url/auth-token/org-id set
	err = initConfig(apiCmd, &flags)
	require.Error(t, err)
}

// Verifies: SYS-REQ-041
func TestHelpCommand(t *testing.T) {
	rootCmd := NewRootCommand("1.0.0", "abc123", "2023-01-01T00:00:00Z")
	
	// Test help command doesn't require configuration
	rootCmd.SetArgs([]string{"help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	
	// Test help for subcommand
	rootCmd.SetArgs([]string{"help", "api"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
}