package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validEnvironment returns a fully-populated Environment that passes Validate.
// reqproof:req REQ-CFG-010
func validEnvironment() *Environment {
	return &Environment{
		Name:           "prod",
		DashboardURL:   "https://dash.example.com",
		AuthToken:      "secret-token",
		OrgID:          "org-123",
		TimeoutSeconds: 30,
	}
}

// reqproof:req REQ-CFG-010
func TestConfig_Validate(t *testing.T) {
	t.Run("no environments returns error", func(t *testing.T) {
		c := &Config{Environments: map[string]*Environment{}}
		err := c.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no environments configured")
	})

	t.Run("no environments (nil map) returns error", func(t *testing.T) {
		c := &Config{}
		err := c.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no environments configured")
	})

	t.Run("no default environment set returns error", func(t *testing.T) {
		c := &Config{
			Environments: map[string]*Environment{
				"prod": validEnvironment(),
			},
		}
		err := c.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no default environment set")
	})

	t.Run("default environment not in map returns error", func(t *testing.T) {
		c := &Config{
			DefaultEnvironment: "staging",
			Environments: map[string]*Environment{
				"prod": validEnvironment(),
			},
		}
		err := c.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "default environment 'staging' not found")
	})

	t.Run("happy path: default exists and is valid", func(t *testing.T) {
		c := &Config{
			DefaultEnvironment: "prod",
			Environments: map[string]*Environment{
				"prod": validEnvironment(),
			},
		}
		err := c.Validate()
		assert.NoError(t, err)
	})

	t.Run("propagates default environment validation error", func(t *testing.T) {
		bad := validEnvironment()
		bad.AuthToken = ""
		c := &Config{
			DefaultEnvironment: "prod",
			Environments: map[string]*Environment{
				"prod": bad,
			},
		}
		err := c.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auth token is required")
	})
}

// reqproof:req REQ-CFG-030
func TestConfig_GetActiveEnvironment(t *testing.T) {
	t.Run("empty default environment returns error", func(t *testing.T) {
		c := &Config{
			Environments: map[string]*Environment{
				"prod": validEnvironment(),
			},
		}
		env, err := c.GetActiveEnvironment()
		require.Error(t, err)
		assert.Nil(t, env)
		assert.Contains(t, err.Error(), "no environments configured or no default environment set")
	})

	t.Run("empty environments map returns error (default set)", func(t *testing.T) {
		// This drives the second condition of the OR independently:
		// DefaultEnvironment is non-empty (F) and Environments is empty (T).
		c := &Config{
			DefaultEnvironment: "prod",
			Environments:       map[string]*Environment{},
		}
		env, err := c.GetActiveEnvironment()
		require.Error(t, err)
		assert.Nil(t, env)
		assert.Contains(t, err.Error(), "no environments configured or no default environment set")
	})

	t.Run("default environment not found in map returns error", func(t *testing.T) {
		c := &Config{
			DefaultEnvironment: "staging",
			Environments: map[string]*Environment{
				"prod": validEnvironment(),
			},
		}
		env, err := c.GetActiveEnvironment()
		require.Error(t, err)
		assert.Nil(t, env)
		assert.Contains(t, err.Error(), "default environment 'staging' not found")
	})

	t.Run("happy path returns environment", func(t *testing.T) {
		want := validEnvironment()
		c := &Config{
			DefaultEnvironment: "prod",
			Environments: map[string]*Environment{
				"prod": want,
			},
		}
		env, err := c.GetActiveEnvironment()
		require.NoError(t, err)
		assert.Same(t, want, env)
	})
}

// reqproof:req REQ-CFG-001
func TestConfig_GetEffectiveConfig(t *testing.T) {
	t.Run("propagates error when no active environment", func(t *testing.T) {
		c := &Config{}
		dash, token, org, err := c.GetEffectiveConfig()
		require.Error(t, err)
		assert.Empty(t, dash)
		assert.Empty(t, token)
		assert.Empty(t, org)
	})

	t.Run("returns dashboard URL, token and org from active environment", func(t *testing.T) {
		c := &Config{
			DefaultEnvironment: "prod",
			Environments: map[string]*Environment{
				"prod": validEnvironment(),
			},
		}
		dash, token, org, err := c.GetEffectiveConfig()
		require.NoError(t, err)
		assert.Equal(t, "https://dash.example.com", dash)
		assert.Equal(t, "secret-token", token)
		assert.Equal(t, "org-123", org)
	})
}

// reqproof:req REQ-CFG-010
func TestEnvironment_Validate(t *testing.T) {
	t.Run("empty name returns error", func(t *testing.T) {
		e := validEnvironment()
		e.Name = ""
		err := e.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "environment name is required")
	})

	t.Run("empty dashboard URL returns error", func(t *testing.T) {
		e := validEnvironment()
		e.DashboardURL = ""
		err := e.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dashboard URL is required")
		assert.Contains(t, err.Error(), e.Name)
	})

	t.Run("dashboard URL missing scheme returns error", func(t *testing.T) {
		// Drives parsedURL.Scheme == "" branch: url.Parse succeeds but Scheme is empty.
		e := validEnvironment()
		e.DashboardURL = "not a url"
		err := e.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dashboard URL format")
	})

	t.Run("dashboard URL missing host returns error", func(t *testing.T) {
		// Drives parsedURL.Host == "" branch: scheme is set but host empty.
		e := validEnvironment()
		e.DashboardURL = "http://"
		err := e.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dashboard URL format")
	})

	t.Run("dashboard URL parse error returns error", func(t *testing.T) {
		// Drives err != nil branch: url.Parse returns an error.
		e := validEnvironment()
		e.DashboardURL = "http://[::1"
		err := e.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dashboard URL format")
	})

	t.Run("empty auth token returns error", func(t *testing.T) {
		e := validEnvironment()
		e.AuthToken = ""
		err := e.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "auth token is required")
		assert.Contains(t, err.Error(), e.Name)
	})

	t.Run("empty org id returns error", func(t *testing.T) {
		e := validEnvironment()
		e.OrgID = ""
		err := e.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID is required")
		assert.Contains(t, err.Error(), e.Name)
	})

	t.Run("all fields set returns nil", func(t *testing.T) {
		e := validEnvironment()
		err := e.Validate()
		assert.NoError(t, err)
	})

	t.Run("error messages include environment name for downstream context", func(t *testing.T) {
		e := validEnvironment()
		e.Name = "staging-eu"
		e.DashboardURL = ""
		err := e.Validate()
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "staging-eu"))
	})
}

// reqproof:req REQ-CFG-010
func TestExitCode_Values(t *testing.T) {
	// Lock the exit code contract — these numeric values are part of the CLI's
	// scriptable surface and changing them is a breaking change.
	assert.Equal(t, ExitCode(0), ExitSuccess)
	assert.Equal(t, ExitCode(1), ExitGeneral)
	assert.Equal(t, ExitCode(2), ExitBadArgs)
	assert.Equal(t, ExitCode(3), ExitNotFound)
	assert.Equal(t, ExitCode(4), ExitConflict)
	assert.Equal(t, ExitCode(5), ExitAuthFailed)
	assert.Equal(t, ExitCode(6), ExitForbidden)
	assert.Equal(t, ExitCode(7), ExitRateLimited)
	assert.Equal(t, ExitCode(8), ExitServerError)
}
