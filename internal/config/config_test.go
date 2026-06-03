package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// reqproof:req REQ-CFG-010
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      types.Config
		expectError bool
	}{
		{
			name: "valid config with environment",
			config: types.Config{
				DefaultEnvironment: "dev",
				Environments: map[string]*types.Environment{
					"dev": {
						Name:         "dev",
						DashboardURL: "http://localhost:3000",
						AuthToken:    "test-token",
						OrgID:        "test-org",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid dashboard URL",
			config: types.Config{
				DefaultEnvironment: "dev",
				Environments: map[string]*types.Environment{
					"dev": {
						Name:         "dev",
						DashboardURL: "invalid-url",
						AuthToken:    "test-token",
						OrgID:        "test-org",
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing auth token",
			config: types.Config{
				DefaultEnvironment: "dev",
				Environments: map[string]*types.Environment{
					"dev": {
						Name:         "dev",
						DashboardURL: "http://localhost:3000",
						OrgID:        "test-org",
					},
				},
			},
			expectError: true,
		},
		{
			name: "missing org ID",
			config: types.Config{
				DefaultEnvironment: "dev",
				Environments: map[string]*types.Environment{
					"dev": {
						Name:         "dev",
						DashboardURL: "http://localhost:3000",
						AuthToken:    "test-token",
					},
				},
			},
			expectError: true,
		},
		{
			name: "no environments configured",
			config: types.Config{
				DefaultEnvironment: "",
				Environments:       make(map[string]*types.Environment),
			},
			expectError: true,
		},
		{
			name: "default environment not found",
			config: types.Config{
				DefaultEnvironment: "prod",
				Environments: map[string]*types.Environment{
					"dev": {
						Name:         "dev",
						DashboardURL: "http://localhost:3000",
						AuthToken:    "test-token",
						OrgID:        "test-org",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// reqproof:req REQ-CFG-030
func TestManagerLoadFromEnvironmentVariables(t *testing.T) {
	// Clean up environment
	originalEnv := map[string]string{
		EnvDashURL:   os.Getenv(EnvDashURL),
		EnvAuthToken: os.Getenv(EnvAuthToken),
		EnvOrgID:     os.Getenv(EnvOrgID),
	}
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set test environment variables
	testDashURL := "http://test-dashboard:3000"
	testAuthToken := "test-auth-token"
	testOrgID := "test-org-id"

	os.Setenv(EnvDashURL, testDashURL)
	os.Setenv(EnvAuthToken, testAuthToken)
	os.Setenv(EnvOrgID, testOrgID)

	manager := NewManager()
	err := manager.LoadConfig()
	assert.NoError(t, err)

	// In unified approach, environment variables should be used to create temporary environments
	// or override existing ones via SetFromFlags
	manager.SetFromFlags(testDashURL, testAuthToken, testOrgID)
	config := manager.GetConfig()
	
	// Check that a temporary environment was created with these values
	assert.NotEmpty(t, config.Environments)
	
	// Get the active environment (should be created by SetFromFlags)
	activeEnv, err := config.GetActiveEnvironment()
	assert.NoError(t, err)
	assert.Equal(t, testDashURL, activeEnv.DashboardURL)
	assert.Equal(t, testAuthToken, activeEnv.AuthToken)
	assert.Equal(t, testOrgID, activeEnv.OrgID)
}

// reqproof:req REQ-CFG-001
func TestManagerFlagsOverrideEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv(EnvDashURL, "http://env-dashboard:3000")
	os.Setenv(EnvAuthToken, "env-auth-token")
	os.Setenv(EnvOrgID, "env-org-id")
	defer func() {
		os.Unsetenv(EnvDashURL)
		os.Unsetenv(EnvAuthToken)
		os.Unsetenv(EnvOrgID)
	}()

	manager := NewManager()
	err := manager.LoadConfig()
	assert.NoError(t, err)

	// Override with flags
	flagDashURL := "http://flag-dashboard:3000"
	flagAuthToken := "flag-auth-token"
	flagOrgID := "flag-org-id"

	manager.SetFromFlags(flagDashURL, flagAuthToken, flagOrgID)

	config := manager.GetConfig()
	activeEnv, err := config.GetActiveEnvironment()
	assert.NoError(t, err)
	
	// Flags should override environment variables
	assert.Equal(t, flagDashURL, activeEnv.DashboardURL)
	assert.Equal(t, flagAuthToken, activeEnv.AuthToken)
	assert.Equal(t, flagOrgID, activeEnv.OrgID)
}

// reqproof:req REQ-CFG-001
func TestManagerPartialFlagOverride(t *testing.T) {
	// Start with an existing environment
	manager := NewManager()
	
	// Create a base environment
	baseEnv := &types.Environment{
		Name:         "dev",
		DashboardURL: "http://base-dashboard:3000",
		AuthToken:    "base-auth-token",
		OrgID:        "base-org-id",
	}
	
	_ = manager.SaveEnvironment(baseEnv, true)

	// Override only dashboard URL with flag
	flagDashURL := "http://flag-dashboard:3000"
	manager.SetFromFlags(flagDashURL, "", "")

	config := manager.GetConfig()
	activeEnv, err := config.GetActiveEnvironment()
	assert.NoError(t, err)
	
	// Only DashboardURL should be overridden
	assert.Equal(t, flagDashURL, activeEnv.DashboardURL)
	assert.Equal(t, "base-auth-token", activeEnv.AuthToken) // Should remain from base
	assert.Equal(t, "base-org-id", activeEnv.OrgID)         // Should remain from base
}

// reqproof:req REQ-CFG-031
func TestManagerEnvironmentOperations(t *testing.T) {
	manager := NewManager()

	// Test saving an environment
	env := &types.Environment{
		Name:         "test",
		DashboardURL: "http://localhost:3000",
		AuthToken:    "test-token",
		OrgID:        "test-org",
	}

	err := manager.SaveEnvironment(env, true)
	assert.NoError(t, err)

	// Test retrieving the environment
	retrieved, err := manager.GetEnvironment("test")
	assert.NoError(t, err)
	assert.Equal(t, env.Name, retrieved.Name)
	assert.Equal(t, env.DashboardURL, retrieved.DashboardURL)
	assert.Equal(t, env.AuthToken, retrieved.AuthToken)
	assert.Equal(t, env.OrgID, retrieved.OrgID)

	// Test that it was set as default
	config := manager.GetConfig()
	assert.Equal(t, "test", config.DefaultEnvironment)

	// Test listing environments
	environments := manager.ListEnvironments()
	assert.Len(t, environments, 1)
	assert.Contains(t, environments, "test")

	// Test setting different default
	env2 := &types.Environment{
		Name:         "prod",
		DashboardURL: "https://prod.example.com",
		AuthToken:    "prod-token",
		OrgID:        "prod-org",
	}

	err = manager.SaveEnvironment(env2, false)
	assert.NoError(t, err)
	
	// Default should still be "test"
	config = manager.GetConfig()
	assert.Equal(t, "test", config.DefaultEnvironment)

	// Switch default
	err = manager.SetDefaultEnvironment("prod")
	assert.NoError(t, err)

	config = manager.GetConfig()
	assert.Equal(t, "prod", config.DefaultEnvironment)
}

// withIsolatedConfigDir points os.UserConfigDir at a fresh temp directory and
// clears the env vars that LoadConfig would otherwise pick up from the
// developer's shell. The returned path is the simulated user-config-dir root
// (i.e. the parent of the "tyk" subdirectory).
// reqproof:req REQ-CFG-030
func withIsolatedConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// On Darwin os.UserConfigDir returns $HOME/Library/Application Support;
	// on Linux it honours $XDG_CONFIG_HOME first, then $HOME/.config. Point
	// both to our temp directory so the lookup is deterministic across OSes.
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	// Make sure no stray env vars from the developer shell leak in and
	// trigger the env-var fallback path unexpectedly.
	t.Setenv(EnvDashURL, "")
	t.Setenv(EnvAuthToken, "")
	t.Setenv(EnvOrgID, "")
	return dir
}

// userConfigTykDir returns the directory LoadConfig actually probes for cli.toml.
// reqproof:req REQ-CFG-030
func userConfigTykDir(t *testing.T) string {
	t.Helper()
	configDir, err := os.UserConfigDir()
	require.NoError(t, err)
	return filepath.Join(configDir, "tyk")
}

// reqproof:req REQ-CFG-030
func TestManagerLoadConfig_EnvVarFallback(t *testing.T) {
	// Each subtest exercises a different combination of env vars driving the
	// "no environments configured" fallback in LoadConfig at line 77:
	//   dashURL != "" || authToken != "" || orgID != ""
	// We need each term independently demonstrated as true with the others
	// false so MC/DC can prove each condition matters.
	tests := []struct {
		name             string
		dashURL          string
		authToken        string
		orgID            string
		wantEnvCreated   bool
		wantDashboardURL string
		wantAuthToken    string
		wantOrgID        string
	}{
		{
			name:           "only dash url set creates default env",
			dashURL:        "http://only-dash:3000",
			wantEnvCreated: true,
			wantDashboardURL: "http://only-dash:3000",
		},
		{
			name:           "only auth token set creates default env",
			authToken:      "only-auth-token",
			wantEnvCreated: true,
			wantAuthToken:  "only-auth-token",
		},
		{
			name:           "only org id set creates default env",
			orgID:          "only-org-id",
			wantEnvCreated: true,
			wantOrgID:      "only-org-id",
		},
		{
			name:           "no env vars set leaves environments empty",
			wantEnvCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withIsolatedConfigDir(t)
			t.Setenv(EnvDashURL, tt.dashURL)
			t.Setenv(EnvAuthToken, tt.authToken)
			t.Setenv(EnvOrgID, tt.orgID)

			manager := NewManager()
			err := manager.LoadConfig()
			require.NoError(t, err)

			cfg := manager.GetConfig()
			if tt.wantEnvCreated {
				require.Contains(t, cfg.Environments, "default")
				env := cfg.Environments["default"]
				assert.Equal(t, tt.wantDashboardURL, env.DashboardURL)
				assert.Equal(t, tt.wantAuthToken, env.AuthToken)
				assert.Equal(t, tt.wantOrgID, env.OrgID)
				assert.Equal(t, "default", cfg.DefaultEnvironment)
			} else {
				assert.Empty(t, cfg.Environments)
			}
		})
	}
}

// reqproof:req REQ-CFG-030
func TestManagerLoadConfig_MalformedConfigFile(t *testing.T) {
	// Force the !ok branch at line 59: ReadInConfig returns a non
	// ConfigFileNotFoundError (parse error), so LoadConfig must surface it.
	withIsolatedConfigDir(t)
	tykDir := userConfigTykDir(t)
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	// Intentionally invalid TOML.
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte("this is = = not valid toml ["), 0o644))

	manager := NewManager()
	err := manager.LoadConfig()
	assert.Error(t, err, "malformed config should bubble up as a non-not-found error")
}

// reqproof:req REQ-CFG-030
func TestManagerLoadConfig_UserConfigDirError(t *testing.T) {
	// Force os.UserConfigDir to fail by unsetting HOME (and XDG vars).
	// LoadConfig should still succeed because the config-file lookup is best
	// effort, and Unmarshal on an empty viper produces an empty config.
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv(EnvDashURL, "")
	t.Setenv(EnvAuthToken, "")
	t.Setenv(EnvOrgID, "")

	manager := NewManager()
	err := manager.LoadConfig()
	require.NoError(t, err)
	assert.Empty(t, manager.GetConfig().Environments)
}

// reqproof:req REQ-CFG-030
func TestManagerLoadConfig_UnmarshalError(t *testing.T) {
	// Drive the err != nil branch at line 66. Inject an incompatible type
	// for the "environments" field directly into viper so mapstructure
	// fails to decode it into map[string]*types.Environment.
	withIsolatedConfigDir(t)

	manager := NewManager()
	manager.GetViperInstance().Set("environments", "definitely-not-a-map")

	err := manager.LoadConfig()
	assert.Error(t, err, "incompatible viper value should bubble up from Unmarshal")
}

// reqproof:req REQ-CFG-030
func TestManagerLoadConfig_ValidConfigFile(t *testing.T) {
	// Drives the err == nil = T && ok = T (file-found) path and the
	// "len(Environments) != 0" branch at line 72 so the env-var fallback is
	// skipped because the config file already populated environments.
	withIsolatedConfigDir(t)
	tykDir := userConfigTykDir(t)
	require.NoError(t, os.MkdirAll(tykDir, 0o755))
	toml := `default_environment = "dev"

[environments.dev]
name = "dev"
dashboard_url = "http://from-file:3000"
auth_token = "file-token"
org_id = "file-org"
`
	require.NoError(t, os.WriteFile(filepath.Join(tykDir, "cli.toml"), []byte(toml), 0o644))

	// Set env vars too; they must NOT create a "default" env because the
	// file already populated environments (len != 0 short-circuits).
	t.Setenv(EnvDashURL, "http://should-be-ignored:9999")

	manager := NewManager()
	require.NoError(t, manager.LoadConfig())

	cfg := manager.GetConfig()
	require.Contains(t, cfg.Environments, "dev")
	assert.NotContains(t, cfg.Environments, "default")
	assert.Equal(t, "dev", cfg.DefaultEnvironment)
	assert.Equal(t, "http://from-file:3000", cfg.Environments["dev"].DashboardURL)
}

// reqproof:req REQ-CFG-001
func TestManagerSetFromFlags_EmptyDashURL(t *testing.T) {
	// Drives line 116 (dashURL != "") to false while still entering
	// SetFromFlags with an active environment available, so the existing
	// DashboardURL is preserved untouched.
	manager := NewManager()
	require.NoError(t, manager.SaveEnvironment(&types.Environment{
		Name:         "dev",
		DashboardURL: "http://base:3000",
		AuthToken:    "base-token",
		OrgID:        "base-org",
	}, true))

	// All flags empty: every if-guard at lines 116, 119, 122 must take the
	// false branch.
	manager.SetFromFlags("", "", "")

	active, err := manager.GetConfig().GetActiveEnvironment()
	require.NoError(t, err)
	assert.Equal(t, "http://base:3000", active.DashboardURL)
	assert.Equal(t, "base-token", active.AuthToken)
	assert.Equal(t, "base-org", active.OrgID)
	// No "temp" env should have been created because we already had an
	// active environment (err == nil at line 108).
	assert.NotContains(t, manager.ListEnvironments(), "temp")
}

// reqproof:req REQ-CFG-001
func TestManagerSetFromFlags_NoActiveEnvCreatesTemp(t *testing.T) {
	// GetActiveEnvironment returns an error (no environments at all), so
	// the line 108 err != nil branch is taken and a "temp" env is created
	// and persisted via SaveEnvironment.
	manager := NewManager()
	manager.SetFromFlags("http://x:3000", "tok", "org")

	active, err := manager.GetConfig().GetActiveEnvironment()
	require.NoError(t, err)
	assert.Equal(t, "temp", active.Name)
	assert.Equal(t, "http://x:3000", active.DashboardURL)
}

// reqproof:req REQ-CFG-031
func TestManagerSaveEnvironment_DefaultEnvironmentBranches(t *testing.T) {
	// Exercises the line 140 short-circuit:
	//   setAsDefault || m.config.DefaultEnvironment == ""
	// Each subtest pins one operand and varies the other.
	t.Run("setAsDefault=false, DefaultEnvironment empty => default gets set", func(t *testing.T) {
		manager := NewManager()
		env := &types.Environment{Name: "first"}
		require.NoError(t, manager.SaveEnvironment(env, false))
		// Because DefaultEnvironment was empty, even with setAsDefault=false
		// the right-hand side of the OR makes it become default.
		assert.Equal(t, "first", manager.GetConfig().DefaultEnvironment)
	})

	t.Run("setAsDefault=false, DefaultEnvironment set => default preserved", func(t *testing.T) {
		manager := NewManager()
		require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "first"}, true))
		// Now DefaultEnvironment == "first", non-empty.
		require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "second"}, false))
		// Both operands of the OR are false: default must NOT change.
		assert.Equal(t, "first", manager.GetConfig().DefaultEnvironment)
	})

	t.Run("setAsDefault=true overrides existing default", func(t *testing.T) {
		manager := NewManager()
		require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "first"}, true))
		require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "second"}, true))
		assert.Equal(t, "second", manager.GetConfig().DefaultEnvironment)
	})
}

// reqproof:req REQ-CFG-031
func TestManagerGetEnvironment_NilEnvironmentsMap(t *testing.T) {
	// A freshly constructed Manager has a Config{} with a nil
	// Environments map: drives the line 149 (Environments == nil) branch to
	// true.
	manager := NewManager()
	require.Nil(t, manager.GetConfig().Environments)

	env, err := manager.GetEnvironment("anything")
	assert.Nil(t, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no environments configured")
}

// reqproof:req REQ-CFG-031
func TestManagerGetEnvironment_NotFound(t *testing.T) {
	// Populate environments so the line 149 (Environments == nil) branch is
	// false, then look up a missing name to drive the line 153 not-found
	// branch.
	manager := NewManager()
	require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "dev"}, true))

	env, err := manager.GetEnvironment("missing")
	assert.Nil(t, env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

// reqproof:req REQ-CFG-031
func TestManagerSetDefaultEnvironment_Branches(t *testing.T) {
	// Drives every combination of the OR at line 171:
	//   m.config.Environments == nil || m.config.Environments[name] == nil
	t.Run("nil environments map errors", func(t *testing.T) {
		manager := NewManager()
		require.Nil(t, manager.GetConfig().Environments)
		err := manager.SetDefaultEnvironment("anything")
		assert.Error(t, err)
	})

	t.Run("environments populated but name missing errors", func(t *testing.T) {
		manager := NewManager()
		require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "dev"}, true))
		err := manager.SetDefaultEnvironment("does-not-exist")
		assert.Error(t, err)
	})

	t.Run("environments populated and name present succeeds", func(t *testing.T) {
		manager := NewManager()
		require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "dev"}, true))
		require.NoError(t, manager.SaveEnvironment(&types.Environment{Name: "prod"}, false))
		require.NoError(t, manager.SetDefaultEnvironment("prod"))
		assert.Equal(t, "prod", manager.GetConfig().DefaultEnvironment)
	})
}

// reqproof:req REQ-CFG-031
func TestManagerListEnvironments_NilMap(t *testing.T) {
	// Ensures the nil-map branch in ListEnvironments returns a fresh empty
	// map rather than nil.
	manager := NewManager()
	require.Nil(t, manager.GetConfig().Environments)
	got := manager.ListEnvironments()
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

// reqproof:req REQ-CFG-030
func TestLiveEnvironmentConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	
	// This test uses the provided live environment details
	testEnv := &types.Environment{
		Name:         "live",
		DashboardURL: "http://tyk-dashboard.localhost:3000",
		AuthToken:    "ff8289874f5d45de945a2ea5c02580fe",
		OrgID:        "5e9d9544a1dcd60001d0ed20",
	}

	// Test that the environment validates correctly
	err := testEnv.Validate()
	assert.NoError(t, err, "Live environment should be valid")

	// Test that we can create a manager and load this environment
	manager := NewManager()
	err = manager.SaveEnvironment(testEnv, true)
	assert.NoError(t, err)
	
	config := manager.GetConfig()
	activeEnv, err := config.GetActiveEnvironment()
	assert.NoError(t, err)
	assert.Equal(t, testEnv.DashboardURL, activeEnv.DashboardURL)
	assert.Equal(t, testEnv.AuthToken, activeEnv.AuthToken)
	assert.Equal(t, testEnv.OrgID, activeEnv.OrgID)

	// Test that configuration validates
	err = config.Validate()
	assert.NoError(t, err, "Configuration with live environment should validate")
}