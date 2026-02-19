package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyktech/tyk-cli/internal/cli"
)

// ---------------------------------------------------------------------------
// Test fixture helpers
// ---------------------------------------------------------------------------

// sampleAPIFile returns a minimal Tyk-enhanced OAS API definition as YAML.
func sampleAPIFile(apiID, name string) string {
	return fmt.Sprintf(`openapi: "3.0.3"
info:
  title: "%s"
  version: "1.0.0"
paths:
  /test:
    get:
      summary: "test"
      responses:
        "200":
          description: "ok"
servers:
  - url: "https://upstream.example.com"
x-tyk-api-gateway:
  info:
    id: "%s"
    name: "%s"
  server:
    listenPath:
      value: "/%s/"
  upstream:
    url: "https://upstream.example.com"
`, name, apiID, name, apiID)
}

// samplePolicyFile returns a minimal policy definition as YAML.
func samplePolicyFile(id, name string) string {
	return fmt.Sprintf(`id: "%s"
name: "%s"
active: true
rate: 1000
per: 60
quota_max: 10000
quota_renewal_rate: 3600
access_rights: {}
`, id, name)
}

// writeFixtureFile writes content to a file inside dir, creating subdirs as needed.
func writeFixtureFile(t *testing.T, dir, relPath, content string) string {
	t.Helper()
	full := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0644))
	return full
}

// createStandardFixtureDir creates a directory with a mix of API + policy files.
// Returns the directory path.
func createStandardFixtureDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	writeFixtureFile(t, dir, "policies/gold-plan.yaml",
		samplePolicyFile("gold-plan", "Gold Plan"))
	writeFixtureFile(t, dir, "policies/silver-plan.yaml",
		samplePolicyFile("silver-plan", "Silver Plan"))
	writeFixtureFile(t, dir, "apis/user-service.yaml",
		sampleAPIFile("user-svc-1", "User Service"))
	writeFixtureFile(t, dir, "apis/payment-service.yaml",
		sampleAPIFile("payment-svc-1", "Payment Service"))

	return dir
}

// requestLog records HTTP requests made to the mock server.
type requestLog struct {
	mu       sync.Mutex
	requests []loggedRequest
}

type loggedRequest struct {
	Method string
	Path   string
	Body   string
}

func (rl *requestLog) add(r *http.Request) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	body := ""
	if r.Body != nil {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r.Body)
		body = buf.String()
	}
	rl.requests = append(rl.requests, loggedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   body,
	})
}

func (rl *requestLog) all() []loggedRequest {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cp := make([]loggedRequest, len(rl.requests))
	copy(cp, rl.requests)
	return cp
}

func (rl *requestLog) count() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.requests)
}

func (rl *requestLog) mutatingCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	n := 0
	for _, r := range rl.requests {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
			n++
		}
	}
	return n
}

// mockDashboard creates an httptest server that simulates the Tyk Dashboard.
// The opts struct controls behavior (which resources exist, whether to fail, etc.).
type mockOpts struct {
	existingAPIs    map[string]bool // apiID -> exists
	existingPolicies map[string]bool // policyID -> exists
	failAPIs        map[string]int  // apiID -> HTTP status to return on apply
	failPolicies    map[string]int  // policyID -> HTTP status to return on apply
	authFailure     bool            // return 401 on all requests
}

func startMockDashboard(t *testing.T, opts mockOpts) (*httptest.Server, *requestLog) {
	t.Helper()
	rlog := &requestLog{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rlog.add(r)

		if opts.authFailure {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Status": "Error",
				"Message": "Unauthorized",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Health endpoint
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`"OK"`))
			return
		}

		// OAS API endpoints
		if strings.Contains(r.URL.Path, "/api/apis/oas") {
			apiID := strings.TrimPrefix(r.URL.Path, "/api/apis/oas/")
			apiID = strings.TrimPrefix(apiID, "/api/apis/oas")
			apiID = strings.TrimPrefix(apiID, "/")

			// Check for forced failure
			if status, fail := opts.failAPIs[apiID]; fail {
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"Status":  "Error",
					"Message": fmt.Sprintf("API apply failed for %s", apiID),
				})
				return
			}

			if r.Method == http.MethodGet {
				if opts.existingAPIs[apiID] {
					_ = json.NewEncoder(w).Encode(map[string]any{
						"openapi": "3.0.3",
						"info":    map[string]any{"title": apiID, "version": "1.0.0"},
						"x-tyk-api-gateway": map[string]any{
							"info": map[string]any{"id": apiID, "name": apiID},
						},
					})
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"Status": "Error", "Message": "API not found",
					})
				}
				return
			}

			if r.Method == http.MethodPost {
				// Create API
				_ = json.NewEncoder(w).Encode(map[string]any{
					"Status": "OK", "Message": "API created", "ID": "new-api-id",
				})
				return
			}

			if r.Method == http.MethodPut || r.Method == http.MethodPatch {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"Status": "OK", "Message": "API updated",
				})
				return
			}
		}

		// Policy endpoints
		if strings.Contains(r.URL.Path, "/api/portal/policies") {
			polID := ""
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			if len(parts) > 3 {
				polID = parts[len(parts)-1]
			}

			// Check for forced failure
			if status, fail := opts.failPolicies[polID]; fail {
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"Status":  "Error",
					"Message": fmt.Sprintf("Policy apply failed for %s", polID),
				})
				return
			}

			if r.Method == http.MethodGet && polID != "" {
				if opts.existingPolicies[polID] {
					_ = json.NewEncoder(w).Encode(map[string]any{
						"_id": polID, "id": polID, "name": polID,
					})
				} else {
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"Status": "Error", "Message": "Policy not found",
					})
				}
				return
			}

			if r.Method == http.MethodGet {
				// List policies
				_ = json.NewEncoder(w).Encode(map[string]any{
					"Data": []any{},
				})
				return
			}

			if r.Method == http.MethodPost {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"Status": "OK", "Message": "Policy created", "Meta": "new-policy-id",
				})
				return
			}

			if r.Method == http.MethodPut {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"Status": "OK", "Message": "Policy updated",
				})
				return
			}
		}

		// Fallback
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": "Error", "Message": "Not found"})
	}))

	t.Cleanup(func() { server.Close() })
	return server, rlog
}

// executeTykApply runs the CLI with the given args against the mock server.
// Returns stdout, stderr, and the cobra exit error (if any).
func executeTykApply(t *testing.T, serverURL string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	// Set env vars for the config system
	t.Setenv("TYK_DASH_URL", serverURL)
	t.Setenv("TYK_AUTH_TOKEN", "test-token")
	t.Setenv("TYK_ORG_ID", "test-org")

	root := cli.NewRootCommand("test", "commit", "time")

	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(args)

	execErr := root.Execute()
	return outBuf.String(), errBuf.String(), execErr
}

// batchResult represents the JSON output of batch apply for assertion.
type batchResult struct {
	Total   int            `json:"total"`
	Applied int            `json:"applied"`
	Failed  int            `json:"failed"`
	Skipped int            `json:"skipped"`
	Results []resultEntry  `json:"results"`
}

type resultEntry struct {
	File      string `json:"file"`
	Type      string `json:"type"`
	Operation string `json:"operation"`
	ID        string `json:"id"`
	Error     string `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Walking Skeleton
// ---------------------------------------------------------------------------

func TestApply_WalkingSkeleton_SingleFileApplied(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "apis/user-service.yaml",
		sampleAPIFile("user-svc-1", "User Service"))

	server, rlog := startMockDashboard(t, mockOpts{})

	stdout, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)
	_ = stdout

	require.NoError(t, err, "apply should succeed; stderr: %s", stderr)
	assert.Contains(t, stderr, "user-service.yaml")
	assert.Contains(t, stderr, "1/1")
	assert.True(t, rlog.count() > 0, "mock server should have received requests")
}

// ---------------------------------------------------------------------------
// US-001: Batch Apply a Directory
// ---------------------------------------------------------------------------

func TestApply_SinglePolicyFile(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "policies/gold-plan.yaml",
		samplePolicyFile("gold-plan", "Gold Plan"))

	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.NoError(t, err, "apply single policy should succeed; stderr: %s", stderr)
	assert.Contains(t, stderr, "gold-plan.yaml")
}

func TestApply_PoliciesAppliedBeforeAPIs(t *testing.T) {
	dir := createStandardFixtureDir(t)
	server, rlog := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.NoError(t, err, "stderr: %s", stderr)

	// Verify ordering: policy requests should appear before API requests
	reqs := rlog.all()
	lastPolicyIdx := -1
	firstAPIIdx := len(reqs)
	for i, r := range reqs {
		if strings.Contains(r.Path, "policies") && (r.Method == http.MethodPost || r.Method == http.MethodPut) {
			if i > lastPolicyIdx {
				lastPolicyIdx = i
			}
		}
		if strings.Contains(r.Path, "/api/apis/oas") && (r.Method == http.MethodPost || r.Method == http.MethodPut) {
			if i < firstAPIIdx {
				firstAPIIdx = i
			}
		}
	}

	if lastPolicyIdx >= 0 && firstAPIIdx < len(reqs) {
		assert.Less(t, lastPolicyIdx, firstAPIIdx,
			"all policy apply requests should come before API apply requests")
	}
}

func TestApply_AllFilesSucceed_ExitZero(t *testing.T) {
	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.NoError(t, err, "all files succeed should exit 0; stderr: %s", stderr)
	// Summary should mention total count
	assert.Contains(t, stderr, "4/4")
}

func TestApply_PartialFailure_ContinuesAndExitOne(t *testing.T) {
	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{
		failAPIs: map[string]int{"payment-svc-1": http.StatusInternalServerError},
	})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.Error(t, err, "partial failure should return error")
	assert.Contains(t, stderr, "FAILED")
	assert.Contains(t, stderr, "payment-service.yaml")
	// Other files should still have been processed
	assert.Contains(t, stderr, "user-service.yaml")
}

func TestApply_EmptyDirectory_ExitTwo(t *testing.T) {
	dir := t.TempDir()
	// Write only non-config files
	writeFixtureFile(t, dir, "README.md", "# This is a readme")

	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.Error(t, err, "empty dir should return error")
	assert.Contains(t, stderr, "No API or policy files found")
}

func TestApply_NonConfigFilesSkipped(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "README.md", "# readme")
	writeFixtureFile(t, dir, "notes.txt", "some notes")
	writeFixtureFile(t, dir, "apis/user-service.yaml",
		sampleAPIFile("user-svc-1", "User Service"))

	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.NoError(t, err, "should succeed ignoring non-config files; stderr: %s", stderr)
	// Only the API file should be mentioned, not .md or .txt
	assert.NotContains(t, stderr, "README.md")
	assert.NotContains(t, stderr, "notes.txt")
	assert.Contains(t, stderr, "user-service.yaml")
}

func TestApply_IdempotentRerun(t *testing.T) {
	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{
		existingAPIs:     map[string]bool{"user-svc-1": true, "payment-svc-1": true},
		existingPolicies: map[string]bool{"gold-plan": true, "silver-plan": true},
	})

	// First run
	_, _, err1 := executeTykApply(t, server.URL, "apply", "-f", dir)
	require.NoError(t, err1)

	// Second run -- should also succeed
	_, stderr2, err2 := executeTykApply(t, server.URL, "apply", "-f", dir)
	require.NoError(t, err2, "idempotent rerun should succeed; stderr: %s", stderr2)
}

func TestApply_SingleFileMode(t *testing.T) {
	dir := t.TempDir()
	apiFile := writeFixtureFile(t, dir, "user-service.yaml",
		sampleAPIFile("user-svc-1", "User Service"))

	server, _ := startMockDashboard(t, mockOpts{})

	// Pass the file directly, not the directory
	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", apiFile)

	require.NoError(t, err, "single file mode should succeed; stderr: %s", stderr)
	assert.Contains(t, stderr, "user-service.yaml")
}

func TestApply_UnrecognizedFileReportedAsFailed(t *testing.T) {
	dir := t.TempDir()
	// Valid YAML but not an API or policy
	writeFixtureFile(t, dir, "mystery.yaml", `name: "something"
version: "1.0"
`)

	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.Error(t, err, "unrecognized file should cause failure")
	assert.Contains(t, stderr, "mystery.yaml")
	assert.Contains(t, stderr, "FAILED")
}

// ---------------------------------------------------------------------------
// US-002: Dry Run
// ---------------------------------------------------------------------------

func TestApply_DryRun_NoMutations(t *testing.T) {

	dir := createStandardFixtureDir(t)
	server, rlog := startMockDashboard(t, mockOpts{
		existingAPIs:     map[string]bool{"user-svc-1": true},
		existingPolicies: map[string]bool{"gold-plan": true},
	})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--dry-run")
	_ = stderr

	require.NoError(t, err)
	assert.Equal(t, 0, rlog.mutatingCount(),
		"dry run should make zero mutating requests")
}

func TestApply_DryRun_ShowsWouldCreateAndWouldUpdate(t *testing.T) {

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{
		existingAPIs:     map[string]bool{"user-svc-1": true},
		existingPolicies: map[string]bool{"gold-plan": true},
	})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--dry-run")

	require.NoError(t, err)
	assert.Contains(t, stderr, "would update")
	assert.Contains(t, stderr, "would create")
}

func TestApply_DryRun_ParseErrorReported(t *testing.T) {

	dir := t.TempDir()
	writeFixtureFile(t, dir, "apis/good.yaml",
		sampleAPIFile("good-api", "Good API"))
	writeFixtureFile(t, dir, "apis/broken.yaml", `this is: [not valid: yaml: "`)

	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--dry-run")
	_ = err // may or may not error depending on implementation

	assert.Contains(t, stderr, "broken.yaml")
	assert.Contains(t, stderr, "FAILED")
}

func TestApply_DryRun_Summary(t *testing.T) {

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--dry-run")

	require.NoError(t, err)
	assert.Contains(t, stderr, "Dry run complete")
	assert.Contains(t, stderr, "0 changes made")
}

// ---------------------------------------------------------------------------
// US-003: Fail-Fast Mode
// ---------------------------------------------------------------------------

func TestApply_FailFast_StopsOnFirstError(t *testing.T) {

	dir := createStandardFixtureDir(t)
	// Fail the first policy alphabetically (gold-plan comes before silver-plan)
	server, rlog := startMockDashboard(t, mockOpts{
		failPolicies: map[string]int{"gold-plan": http.StatusInternalServerError},
	})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--continue-on-error=false")

	require.Error(t, err)
	assert.Contains(t, stderr, "FAILED")
	// Should not have processed all 4 files
	mutating := rlog.mutatingCount()
	assert.Less(t, mutating, 4, "fail-fast should stop before processing all files")
}

func TestApply_FailFast_RemainingSkipped(t *testing.T) {

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{
		failPolicies: map[string]int{"gold-plan": http.StatusInternalServerError},
	})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--continue-on-error=false")

	require.Error(t, err)
	assert.Contains(t, stderr, "skipped")
}

func TestApply_AuthFailure_AlwaysStops_ExitThree(t *testing.T) {

	dir := createStandardFixtureDir(t)
	server, rlog := startMockDashboard(t, mockOpts{
		authFailure: true,
	})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.Error(t, err)
	assert.Contains(t, stderr, "Authentication failed")
	// Should have stopped after first request
	assert.LessOrEqual(t, rlog.mutatingCount(), 1,
		"auth failure should stop after first attempt")
}

func TestApply_FailFast_AllSucceed_SameAsDefault(t *testing.T) {

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--continue-on-error=false")

	require.NoError(t, err, "all succeed in fail-fast should exit 0; stderr: %s", stderr)
	assert.Contains(t, stderr, "4/4")
}

// ---------------------------------------------------------------------------
// US-004: JSON Output
// ---------------------------------------------------------------------------

func TestApply_JSONOutput_ValidStructure(t *testing.T) {
	t.Skip("not yet implemented")

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{})

	stdout, _, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--json")

	require.NoError(t, err)

	var result batchResult
	require.NoError(t, json.Unmarshal([]byte(stdout), &result),
		"stdout should be valid JSON: %s", stdout)

	assert.Equal(t, 4, result.Total)
	assert.Equal(t, 4, result.Applied)
	assert.Equal(t, 0, result.Failed)
	assert.Len(t, result.Results, 4)

	for _, r := range result.Results {
		assert.NotEmpty(t, r.File, "each result should have a file path")
		assert.NotEmpty(t, r.Type, "each result should have a type")
		assert.NotEmpty(t, r.Operation, "each result should have an operation")
		assert.Contains(t, []string{"api", "policy"}, r.Type)
		assert.Contains(t, []string{"created", "updated", "unchanged"}, r.Operation)
	}
}

func TestApply_JSONOutput_FailedEntryHasError(t *testing.T) {
	t.Skip("not yet implemented")

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{
		failAPIs: map[string]int{"payment-svc-1": http.StatusInternalServerError},
	})

	stdout, _, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--json")
	_ = err // exit code 1 expected

	var result batchResult
	require.NoError(t, json.Unmarshal([]byte(stdout), &result),
		"stdout should be valid JSON even on failure: %s", stdout)

	assert.Equal(t, 1, result.Failed)

	foundFailed := false
	for _, r := range result.Results {
		if r.Operation == "failed" {
			foundFailed = true
			assert.NotEmpty(t, r.Error, "failed entry should have an error field")
			assert.Contains(t, r.File, "payment-service.yaml")
		}
	}
	assert.True(t, foundFailed, "should have at least one failed entry")
}

func TestApply_JSONOutput_DryRun(t *testing.T) {
	t.Skip("not yet implemented")

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{
		existingAPIs:     map[string]bool{"user-svc-1": true},
		existingPolicies: map[string]bool{"gold-plan": true},
	})

	stdout, _, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--dry-run", "--json")

	require.NoError(t, err)

	var result batchResult
	require.NoError(t, json.Unmarshal([]byte(stdout), &result),
		"JSON dry-run output should be valid: %s", stdout)

	assert.Equal(t, 0, result.Applied, "dry run should report 0 applied")

	hasWouldCreate := false
	hasWouldUpdate := false
	for _, r := range result.Results {
		if r.Operation == "would_create" {
			hasWouldCreate = true
		}
		if r.Operation == "would_update" {
			hasWouldUpdate = true
		}
	}
	assert.True(t, hasWouldCreate, "should have would_create entries")
	assert.True(t, hasWouldUpdate, "should have would_update entries")
}

func TestApply_JSONOutput_FailFast_IncludesSkipped(t *testing.T) {
	t.Skip("not yet implemented")

	dir := createStandardFixtureDir(t)
	server, _ := startMockDashboard(t, mockOpts{
		failPolicies: map[string]int{"gold-plan": http.StatusInternalServerError},
	})

	stdout, _, err := executeTykApply(t, server.URL, "apply", "-f", dir,
		"--continue-on-error=false", "--json")
	_ = err

	var result batchResult
	require.NoError(t, json.Unmarshal([]byte(stdout), &result),
		"JSON fail-fast output should be valid: %s", stdout)

	assert.Greater(t, result.Skipped, 0, "should have skipped entries")

	hasSkipped := false
	for _, r := range result.Results {
		if r.Operation == "skipped" {
			hasSkipped = true
		}
	}
	assert.True(t, hasSkipped, "results array should contain skipped entries")
}
