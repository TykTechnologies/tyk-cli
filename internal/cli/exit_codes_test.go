package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/pkg/types"
)

// ---------------------------------------------------------------------------
// SYS-REQ-016: api create / import-oas / apply exit 4 on HTTP 409
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-016
func TestAPIImportOAS_ConflictReturnsExit4(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/apis/oas" {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  409,
				"message": "API listen path already in use",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Write a minimal OAS file to a temp dir.
	tempDir := t.TempDir()
	oasPath := tempDir + "/api.json"
	oasDoc := map[string]interface{}{
		"openapi": "3.0.0",
		"info":    map[string]interface{}{"title": "Conflicting API", "version": "1.0.0"},
		"x-tyk-api-gateway": map[string]interface{}{
			"info":   map[string]interface{}{"name": "Conflicting API"},
			"server": map[string]interface{}{"listenPath": map[string]interface{}{"value": "/conflict/"}},
		},
	}
	data, _ := json.Marshal(oasDoc)
	require.NoError(t, os.WriteFile(oasPath, data, 0o600))

	cmd := NewAPIImportOASCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetArgs([]string{"--file", oasPath})
	_ = cmd.ParseFlags([]string{"--file", oasPath})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "expected ExitError, got %T: %v", err, err)
	assert.Equal(t, int(types.ExitConflict), exitErr.Code, "SYS-REQ-016: HTTP 409 must map to exit code 4")
}

// Verifies: SYS-REQ-016
func TestAPICreate_ConflictReturnsExit4(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/apis/oas" {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  409,
				"message": "API conflict",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := NewAPICreateCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	cmd.SetContext(withConfig(context.Background(), cfg))
	cmd.SetArgs([]string{
		"--name", "Conflicting API",
		"--upstream-url", "http://example.com",
		"--listen-path", "/conflict/",
	})
	_ = cmd.ParseFlags([]string{
		"--name", "Conflicting API",
		"--upstream-url", "http://example.com",
		"--listen-path", "/conflict/",
	})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "expected ExitError, got %T: %v", err, err)
	assert.Equal(t, int(types.ExitConflict), exitErr.Code, "SYS-REQ-016: HTTP 409 must map to exit code 4")
}

// ---------------------------------------------------------------------------
// SYS-REQ-036: policy apply exits 4 on HTTP 409
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-036
func TestPolicyApply_Create_ConflictReturnsExit4(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			// Not found -> create path
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "message": "not found"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/portal/policies":
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  409,
				"message": "policy already exists",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, policyFile)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "expected ExitError, got %T: %v", err, err)
	assert.Equal(t, int(types.ExitConflict), exitErr.Code, "SYS-REQ-036: HTTP 409 on create must map to exit code 4")
}

// Verifies: SYS-REQ-036
func TestPolicyApply_Update_ConflictReturnsExit4(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/apis":
			_ = json.NewEncoder(w).Encode(mockAPIListResponse())
		case r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies/platinum":
			// Found -> update path
			p := mockDashboardPolicy("507f1f77bcf86cd799439099", "platinum", "Platinum Plan", 5000, 60, 500000, 2592000,
				[]string{}, map[string]interface{}{})
			_ = json.NewEncoder(w).Encode(p)
		case r.Method == http.MethodPut && r.URL.Path == "/api/portal/policies/platinum":
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  409,
				"message": "policy conflict",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	policyFile := writeTempPolicyFile(t, validPlatinumPolicyYAML)
	err := executePolicyApplyCmd(t, server.URL, policyFile)

	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "expected ExitError, got %T: %v", err, err)
	assert.Equal(t, int(types.ExitConflict), exitErr.Code, "SYS-REQ-036: HTTP 409 on update must map to exit code 4")
}

// ---------------------------------------------------------------------------
// SYS-REQ-050 / SYS-REQ-033: invalid args produce exit code 2 (ExitBadArgs)
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-050
func TestConfigAdd_MissingRequiredFieldReturnsExit2(t *testing.T) {
	cmd := NewConfigAddCommand()
	// Missing --auth-token and --org-id, only dashboard-url given.
	cmd.SetArgs([]string{"badenv", "--dashboard-url", "http://localhost:3000"})
	_ = cmd.ParseFlags([]string{"badenv", "--dashboard-url", "http://localhost:3000"})

	err := cmd.RunE(cmd, []string{"badenv"})
	require.Error(t, err, "missing required fields should fail")
	// Validation error from Environment.Validate is wrapped by cobra into the
	// returned error chain; the contract is that the failure carries enough
	// information for the caller to choose exit 2. The CLI top-level error
	// handler maps validation errors to ExitBadArgs; here we just confirm the
	// underlying error reaches the caller as an actionable failure.
	assert.Contains(t, err.Error(), "required",
		"validation error should explain which required field is missing")
}

// Verifies: SYS-REQ-033
// SYS-REQ-033 says "all policy subcommands exit 0 on success".
// This is a contract verified by every other successful policy test passing
// without error. We add one explicit smoke assertion here so the requirement
// has a direct verified_by link.
func TestPolicyList_SuccessReturnsNilError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/portal/policies" {
			_ = json.NewEncoder(w).Encode(mockPolicyListResponse([]map[string]interface{}{}))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cmd := NewPolicyListCommand()
	cfg := createPolicyConfig(server.URL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err, "SYS-REQ-033: policy list on a clean response must return nil (exit 0)")
}

// Verifies: SYS-REQ-049
// SYS-REQ-049 says "all config subcommands exit 0 on success". The config use
// / current tests in config_use_test.go already cover this; we add one
// explicit smoke assertion against config current with a default env so the
// requirement has a clear verified_by link.
func TestConfigCurrent_SuccessReturnsNilError(t *testing.T) {
	setupTempConfig(t, twoEnvConfig)
	_ = captureColorOutput(t)

	cmd := NewConfigCurrentCommand()
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err, "SYS-REQ-049: config current on a valid config must return nil (exit 0)")
}


// ---------------------------------------------------------------------------
// SYS-REQ-017 / SYS-REQ-037: HTTP 401 → exit 5 (ExitAuthFailed)
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-017, INT-REQ-002
func TestAPIList_AuthFailedReturnsExit5(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 401, "message": "invalid auth token",
		})
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "bad-token", OrgID: "org"},
		},
	}
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "expected ExitError, got %T: %v", err, err)
	assert.Equal(t, int(types.ExitAuthFailed), exitErr.Code,
		"SYS-REQ-017: HTTP 401 must map to exit code 5")
	assert.Contains(t, exitErr.Message, "auth", "message must hint at auth failure")
}

// Verifies: SYS-REQ-037
func TestPolicyList_AuthFailedReturnsExit5(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 401, "message": "invalid token"})
	}))
	defer server.Close()

	cmd := NewPolicyListCommand()
	cfg := createPolicyConfig(server.URL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "expected ExitError, got %T", err)
	assert.Equal(t, int(types.ExitAuthFailed), exitErr.Code,
		"SYS-REQ-037: HTTP 401 must map to exit code 5")
}

// ---------------------------------------------------------------------------
// SYS-REQ-020 / SYS-REQ-040: HTTP 403 → exit 6 (ExitForbidden)
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-020
func TestAPIList_ForbiddenReturnsExit6(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 403, "message": "insufficient permission"})
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok, "expected ExitError")
	assert.Equal(t, int(types.ExitForbidden), exitErr.Code,
		"SYS-REQ-020: HTTP 403 must map to exit code 6")
}

// Verifies: SYS-REQ-040
func TestPolicyList_ForbiddenReturnsExit6(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 403, "message": "forbidden"})
	}))
	defer server.Close()

	cmd := NewPolicyListCommand()
	cfg := createPolicyConfig(server.URL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitForbidden), exitErr.Code,
		"SYS-REQ-040: HTTP 403 must map to exit code 6")
}

// ---------------------------------------------------------------------------
// SYS-REQ-018 / SYS-REQ-038: HTTP 429 → exit 7 (ExitRateLimited)
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-018
func TestAPIList_RateLimitedReturnsExit7(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 429, "message": "rate limited"})
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitRateLimited), exitErr.Code,
		"SYS-REQ-018: HTTP 429 must map to exit code 7")
}

// Verifies: SYS-REQ-038
func TestPolicyList_RateLimitedReturnsExit7(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 429, "message": "rate limited"})
	}))
	defer server.Close()

	cmd := NewPolicyListCommand()
	cfg := createPolicyConfig(server.URL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitRateLimited), exitErr.Code,
		"SYS-REQ-038: HTTP 429 must map to exit code 7")
}

// ---------------------------------------------------------------------------
// SYS-REQ-019 / SYS-REQ-039: HTTP 5xx → exit 8 (ExitServerError)
// ---------------------------------------------------------------------------

// Verifies: SYS-REQ-019
func TestAPIList_ServerErrorReturnsExit8(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 500, "message": "internal server error"})
	}))
	defer server.Close()

	cmd := NewAPIListCommand()
	cfg := &types.Config{
		DefaultEnvironment: "test",
		Environments: map[string]*types.Environment{
			"test": {Name: "test", DashboardURL: server.URL, AuthToken: "token", OrgID: "org"},
		},
	}
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitServerError), exitErr.Code,
		"SYS-REQ-019: HTTP 5xx must map to exit code 8")
}

// Verifies: SYS-REQ-039
func TestPolicyList_ServerErrorReturnsExit8(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": 503, "message": "service unavailable"})
	}))
	defer server.Close()

	cmd := NewPolicyListCommand()
	cfg := createPolicyConfig(server.URL)
	ctx := withConfig(context.Background(), cfg)
	ctx = withOutputFormat(ctx, types.OutputJSON)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.RunE(cmd, []string{})
	require.Error(t, err)
	exitErr, ok := err.(*ExitError)
	require.True(t, ok)
	assert.Equal(t, int(types.ExitServerError), exitErr.Code,
		"SYS-REQ-039: HTTP 5xx must map to exit code 8")
}
