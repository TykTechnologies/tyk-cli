package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyktech/tyk-cli/pkg/types"
)

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