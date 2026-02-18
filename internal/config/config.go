package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/tyktech/tyk-cli/pkg/types"
)

const (
	// Environment variable names
	EnvDashURL   = "TYK_DASH_URL"
	EnvAuthToken = "TYK_AUTH_TOKEN"
	EnvOrgID     = "TYK_ORG_ID"

	// Config file name (without extension)
	ConfigFileName = "cli"
	ConfigFileType = "toml"
)

// Manager handles configuration loading and management
type Manager struct {
	viper  *viper.Viper
	config *types.Config
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	v := viper.New()
	
	// Set environment variable names and automatic env binding
	v.SetEnvPrefix("TYK")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values for unified environment system
	v.SetDefault("default_environment", "")

	return &Manager{
		viper: v,
		config: &types.Config{},
	}
}

// LoadConfig loads configuration from environment, config file, and flags
func (m *Manager) LoadConfig() error {
	// Try to load config file from user config directory
	if configDir, err := os.UserConfigDir(); err == nil {
		tykConfigDir := filepath.Join(configDir, "tyk")
		m.viper.AddConfigPath(tykConfigDir)
		m.viper.SetConfigName(ConfigFileName)
		m.viper.SetConfigType(ConfigFileType)

		// Read config file if it exists (ignore errors if file doesn't exist)
		if err := m.viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}
	}

	// Unmarshal into our config struct
	if err := m.viper.Unmarshal(m.config); err != nil {
		return err
	}

	// If no environments are configured, check for individual environment variables
	// and create a default environment from them
	if len(m.config.Environments) == 0 {
		dashURL := m.viper.GetString("dash_url")
		authToken := m.viper.GetString("auth_token")
		orgID := m.viper.GetString("org_id")

		if dashURL != "" || authToken != "" || orgID != "" {
			// Create default environment from environment variables
			env := &types.Environment{
				Name:         "default",
				DashboardURL: dashURL,
				AuthToken:    authToken,
				OrgID:        orgID,
			}
			_ = m.SaveEnvironment(env, true)
		}
	}

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *types.Config {
	return m.config
}

// GetEffectiveConfig returns a config with values resolved from the active environment
func (m *Manager) GetEffectiveConfig() *types.Config {
	// In unified approach, just return the config as-is
	// The active environment is accessed via GetActiveEnvironment()
	return m.config
}

// SetFromFlags updates the current environment with values from command line flags
func (m *Manager) SetFromFlags(dashURL, authToken, orgID string) {
	// Get or create a temporary environment for flag overrides
	activeEnv, err := m.config.GetActiveEnvironment()
	if err != nil {
		// If no environments exist, create a temporary one for compatibility
		activeEnv = &types.Environment{
			Name: "temp",
		}
	}
	
	// Apply flag overrides to active environment
	if dashURL != "" {
		activeEnv.DashboardURL = dashURL
	}
	if authToken != "" {
		activeEnv.AuthToken = authToken
	}
	if orgID != "" {
		activeEnv.OrgID = orgID
	}
	
	// If we had to create a temp environment, save it
	if activeEnv.Name == "temp" {
		_ = m.SaveEnvironment(activeEnv, true)
	}
}

// SaveEnvironment saves an environment to the configuration
func (m *Manager) SaveEnvironment(env *types.Environment, setAsDefault bool) error {
	if m.config.Environments == nil {
		m.config.Environments = make(map[string]*types.Environment)
	}
	
	m.config.Environments[env.Name] = env
	
	if setAsDefault || m.config.DefaultEnvironment == "" {
		m.config.DefaultEnvironment = env.Name
	}
	
	return nil
}

// GetEnvironment returns a specific environment
func (m *Manager) GetEnvironment(name string) (*types.Environment, error) {
	if m.config.Environments == nil {
		return nil, fmt.Errorf("no environments configured")
	}
	
	env, exists := m.config.Environments[name]
	if !exists {
		return nil, fmt.Errorf("environment '%s' not found", name)
	}
	
	return env, nil
}

// ListEnvironments returns all configured environments
func (m *Manager) ListEnvironments() map[string]*types.Environment {
	if m.config.Environments == nil {
		return make(map[string]*types.Environment)
	}
	return m.config.Environments
}

// SetDefaultEnvironment sets the default environment
func (m *Manager) SetDefaultEnvironment(name string) error {
	if m.config.Environments == nil || m.config.Environments[name] == nil {
		return fmt.Errorf("environment '%s' not found", name)
	}
	
	m.config.DefaultEnvironment = name
	return nil
}

// GetViperInstance returns the underlying viper instance for testing
func (m *Manager) GetViperInstance() *viper.Viper {
	return m.viper
}