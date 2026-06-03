package types

import (
	"errors"
	"fmt"
	"net/url"
)

// Config holds all configuration for the Tyk CLI
// In the unified approach, config IS environments - no base config fields
type Config struct {
	// Default active environment
	DefaultEnvironment string                   `mapstructure:"default_environment" yaml:"default_environment" json:"default_environment"`
	// All named environments (this IS the configuration system)
	Environments       map[string]*Environment  `mapstructure:"environments" yaml:"environments" json:"environments"`
}

// Environment represents a named configuration environment
// In the unified model, environments ARE the configuration
type Environment struct {
	Name           string `mapstructure:"name"            yaml:"name"            json:"name"`
	DashboardURL   string `mapstructure:"dashboard_url"   yaml:"dashboard_url"   json:"dashboard_url"`
	AuthToken      string `mapstructure:"auth_token"      yaml:"auth_token"      json:"auth_token"`
	OrgID          string `mapstructure:"org_id"          yaml:"org_id"          json:"org_id"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds" yaml:"timeout_seconds" json:"timeout_seconds,omitempty"`
}

// Validate checks if the configuration is valid
// reqproof:req REQ-CFG-010
func (c *Config) Validate() error {
	// Must have at least one environment
	if len(c.Environments) == 0 {
		return errors.New("no environments configured. Use 'tyk config add' to add an environment")
	}

	// Must have a default environment set
	if c.DefaultEnvironment == "" {
		return errors.New("no default environment set")
	}

	// Validate that the default environment exists
	env, exists := c.Environments[c.DefaultEnvironment]
	if !exists {
		return fmt.Errorf("default environment '%s' not found", c.DefaultEnvironment)
	}

	// Validate the default environment
	return env.Validate()
}

// reqproof:req REQ-CFG-030
func (c *Config) GetActiveEnvironment() (*Environment, error) {
	if c.DefaultEnvironment == "" || len(c.Environments) == 0 {
		return nil, errors.New("no environments configured or no default environment set")
	}

	env, exists := c.Environments[c.DefaultEnvironment]
	if !exists {
		return nil, fmt.Errorf("default environment '%s' not found", c.DefaultEnvironment)
	}

	return env, nil
}

// reqproof:req REQ-CFG-001
func (c *Config) GetEffectiveConfig() (string, string, string, error) {
	env, err := c.GetActiveEnvironment()
	if err != nil {
		return "", "", "", err
	}
	return env.DashboardURL, env.AuthToken, env.OrgID, nil
}

// reqproof:req REQ-CFG-010
func (e *Environment) Validate() error {
	if e.Name == "" {
		return errors.New("environment name is required")
	}

	if e.DashboardURL == "" {
		return fmt.Errorf("dashboard URL is required for environment '%s'", e.Name)
	}

	// Validate URL format
	parsedURL, err := url.Parse(e.DashboardURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("invalid dashboard URL format for environment '%s': %s", e.Name, e.DashboardURL)
	}

	if e.AuthToken == "" {
		return fmt.Errorf("auth token is required for environment '%s'", e.Name)
	}

	if e.OrgID == "" {
		return fmt.Errorf("organization ID is required for environment '%s'", e.Name)
	}

	return nil
}

// ExitCode represents different types of CLI exit codes
//
// reqproof:req REQ-API-020
// reqproof:req REQ-API-021
// reqproof:req REQ-CFG-020
// reqproof:req REQ-CFG-021
// reqproof:req REQ-POL-020
// reqproof:req REQ-POL-021
type ExitCode int

const (
	ExitSuccess     ExitCode = 0 // Success
	ExitGeneral     ExitCode = 1 // Generic failure (I/O, network, unexpected)
	ExitBadArgs     ExitCode = 2 // Bad arguments (missing file, invalid flag combination)
	ExitNotFound    ExitCode = 3 // Not found (API or version)
	ExitConflict    ExitCode = 4 // Conflict (e.g. creating an API that already exists without --force)
	ExitAuthFailed  ExitCode = 5 // Authentication failed (HTTP 401 — invalid or expired token)
	ExitForbidden   ExitCode = 6 // Forbidden (HTTP 403 — token lacks permission)
	ExitRateLimited ExitCode = 7 // Rate-limited by Dashboard (HTTP 429)
	ExitServerError ExitCode = 8 // Dashboard server error (HTTP 5xx after retry attempts)
)

// OutputFormat represents the output format for CLI commands
type OutputFormat string

const (
	OutputHuman OutputFormat = "human"
	OutputJSON  OutputFormat = "json"
)