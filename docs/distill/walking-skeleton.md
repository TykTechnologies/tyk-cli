# Walking Skeleton: OAS Versioning

## First Test to Implement

**WS-01: List versions for a multi-version API** (scenario S-01.1 / US-01)

This is the walking skeleton because:

1. It exercises the simplest end-to-end slice: user runs command, CLI calls Dashboard API, CLI formats output
2. It validates the driving port pattern (cobra command RunE with injected config and httptest mock)
3. It uses an existing client method (`ListOASAPIVersions`) so no new production code in the client layer
4. All other commands depend on list-versions (create needs it for conflict check, switch-default needs it for validation)
5. It is demo-able: "can a user see their API versions?" is a clear user goal with observable output

## Walking Skeleton Implementation Sequence

### Skeleton 1: WS-01 -- List versions (human output)

**Test file**: `internal/cli/api_versions_test.go`

**What to build to make it pass**:
- Replace placeholder `NewAPIVersionsListCommand()` in `internal/cli/api.go`
- Wire `client.ListOASAPIVersions()` call
- Format version table to stderr

**Mock server shape**:
```go
// GET /api/apis/oas/{apiID}/versions -> VersionListResponse
mux.HandleFunc("/api/apis/oas/abc123/versions", func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]any{
        "apis": []any{
            map[string]any{"api_id": "abc123-v1", "version_name": "v1"},
            map[string]any{"api_id": "abc123-v2", "version_name": "v2"},
            map[string]any{"api_id": "abc123-v3", "version_name": "v3"},
        },
        "default_version": "v1",
    })
})
```

**Test pattern** (matches `executePolicyListCmd` in `policy_test.go`):
```go
func executeVersionsListCmd(t *testing.T, serverURL string, format types.OutputFormat, apiID string) (error) {
    root := NewRootCommand("test", "commit", "time")
    cmd, _, err := root.Find([]string{"api", "versions", "list"})
    require.NoError(t, err)

    cfg := createVersionConfig(serverURL)
    ctx := withConfig(context.Background(), cfg)
    ctx = withOutputFormat(ctx, format)
    cmd.SetContext(ctx)
    cmd.SetArgs([]string{"--api-id", apiID})
    _ = cmd.ParseFlags([]string{"--api-id", apiID})

    return cmd.RunE(cmd, []string{})
}
```

### Skeleton 2: WS-02 -- Create version from OAS file

**Test file**: `internal/cli/api_versions_test.go`

**What to build**: Replace placeholder `NewAPIVersionsCreateCommand()`, add `client.CreateOASAPIVersion()`

**Depends on**: WS-01 passing (list is used for conflict check)

### Skeleton 3: WS-03 -- Switch default version

**Test file**: `internal/cli/api_versions_test.go`

**What to build**: Replace placeholder `NewAPIVersionsSwitchDefaultCommand()`

**Depends on**: WS-01 passing (list is used for validation)

## Implementation Order (One Test at a Time)

All tests start as `t.Skip("not yet implemented")`. Enable one, make it pass, commit.

| Order | Scenario ID | Description | Skip Until |
|-------|------------|-------------|------------|
| 1 | WS-01 | List versions (human output) | Enable first |
| 2 | S-01.2 | List versions (JSON output) | WS-01 passes |
| 3 | S-01.4 | List versions (API not found, exit 3) | S-01.2 passes |
| 4 | S-03.1 | Switch default version | S-01.4 passes |
| 5 | S-03.3 | Switch default (no-op, already default) | S-03.1 passes |
| 6 | S-03.4 | Switch default (version not found, exit 3) | S-03.3 passes |
| 7 | S-02.1 | Create version from OAS file | S-03.4 passes |
| 8 | S-02.4 | Create version (duplicate, exit 4) | S-02.1 passes |
| 9 | S-02.5 | Create version (invalid OAS, exit 2) | S-02.4 passes |
| 10 | S-04.1 | Single file apply as version | S-02.5 passes |
| 11 | S-04.2 | Batch apply discovers version files | S-04.1 passes |
| 12 | S-04.4 | Dry-run reports version operations | S-04.2 passes |
| Remaining | All other scenarios | Fill in coverage | After core path works |

## Test File Location

```
internal/cli/api_versions_test.go    -- US-01, US-02, US-03, US-05 (unit-level, in cli package)
test/apply_version_integration_test.go -- US-04 (integration-level, matches existing apply test pattern)
```
