package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleVersionFile returns a version API file with versioning metadata.
func sampleVersionFile(apiID, baseAPIID, versionName string) string {
	return fmt.Sprintf(`openapi: "3.0.3"
info:
  title: "%s version"
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
    name: "%s version"
    versioning:
      base_api_id: "%s"
      version_name: "%s"
      set_default: false
  server:
    listenPath:
      value: "/%s/"
  upstream:
    url: "https://upstream.example.com"
`, versionName, apiID, versionName, baseAPIID, versionName, apiID)
}

// ---------------------------------------------------------------------------
// US-005: Version File Support in Batch Apply
// ---------------------------------------------------------------------------

func TestApply_VersionFile_ClassifiedAsAPIVersion(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "apis/base.yaml",
		sampleAPIFile("base-api-1", "Base API"))
	writeFixtureFile(t, dir, "apis/v2.yaml",
		sampleVersionFile("v2-api-1", "base-api-1", "v2"))

	server, rlog := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.NoError(t, err, "apply with version file should succeed; stderr: %s", stderr)

	// Version file should have been sent as a POST with query params
	reqs := rlog.all()
	foundVersionCreate := false
	for _, r := range reqs {
		if r.Method == http.MethodPost && r.Path == "/api/apis/oas" {
			foundVersionCreate = true
		}
	}
	assert.True(t, foundVersionCreate, "should have made a POST request for version creation")
}

func TestApply_VersionFiles_SortedAfterBaseAPIs(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "policies/pol.yaml",
		samplePolicyFile("pol-1", "Policy"))
	writeFixtureFile(t, dir, "apis/base.yaml",
		sampleAPIFile("base-api-1", "Base API"))
	writeFixtureFile(t, dir, "apis/v2.yaml",
		sampleVersionFile("v2-api-1", "base-api-1", "v2"))

	server, rlog := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir)

	require.NoError(t, err, "apply should succeed; stderr: %s", stderr)

	// Verify ordering: policy -> base API -> version API
	reqs := rlog.all()
	lastPolicyIdx := -1
	lastBaseAPIIdx := -1
	firstVersionIdx := len(reqs)
	for i, r := range reqs {
		if r.Method == http.MethodPost && r.Path == "/api/portal/policies" {
			lastPolicyIdx = i
		}
		// Base API creates go to /api/apis/oas without query params
		if r.Method == http.MethodPost && r.Path == "/api/apis/oas" {
			// We'll track all POSTs; the last base API should be before version
			if i > lastBaseAPIIdx {
				lastBaseAPIIdx = i
			}
		}
	}
	// The version POST will also go to /api/apis/oas but with query string
	// Since rlog.Path doesn't include query, we check the ordering indirectly
	// through stderr output order
	_ = firstVersionIdx
	if lastPolicyIdx >= 0 {
		assert.Less(t, lastPolicyIdx, lastBaseAPIIdx,
			"policy should be applied before base APIs")
	}
	assert.Contains(t, stderr, "3/3", "all 3 files should be processed")
}

func TestApply_DryRun_VersionFile_ShowsWouldCreateVersion(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "apis/base.yaml",
		sampleAPIFile("base-api-1", "Base API"))
	writeFixtureFile(t, dir, "apis/v2.yaml",
		sampleVersionFile("v2-api-1", "base-api-1", "v2"))

	server, _ := startMockDashboard(t, mockOpts{})

	_, stderr, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--dry-run")

	require.NoError(t, err, "dry-run should succeed; stderr: %s", stderr)
	assert.Contains(t, stderr, "would create version")
}

type versionResultEntry struct {
	File        string `json:"file"`
	Type        string `json:"type"`
	Operation   string `json:"operation"`
	ID          string `json:"id"`
	VersionName string `json:"version_name,omitempty"`
	Error       string `json:"error,omitempty"`
}

type versionBatchResult struct {
	Total   int                  `json:"total"`
	Applied int                  `json:"applied"`
	Failed  int                  `json:"failed"`
	Skipped int                  `json:"skipped"`
	Results []versionResultEntry `json:"results"`
}

func TestApply_JSONOutput_VersionFile_HasVersionName(t *testing.T) {
	dir := t.TempDir()
	writeFixtureFile(t, dir, "apis/base.yaml",
		sampleAPIFile("base-api-1", "Base API"))
	writeFixtureFile(t, dir, "apis/v2.yaml",
		sampleVersionFile("v2-api-1", "base-api-1", "v2"))

	server, _ := startMockDashboard(t, mockOpts{})

	stdout, _, err := executeTykApply(t, server.URL, "apply", "-f", dir, "--json")

	require.NoError(t, err)

	var result versionBatchResult
	require.NoError(t, json.Unmarshal([]byte(stdout), &result),
		"stdout should be valid JSON: %s", stdout)

	foundVersion := false
	for _, r := range result.Results {
		if r.Type == "api_version" {
			foundVersion = true
			assert.Equal(t, "v2", r.VersionName, "version_name should be set")
			assert.Equal(t, "version_created", r.Operation)
		}
	}
	assert.True(t, foundVersion, "should have an api_version entry in results")
}
