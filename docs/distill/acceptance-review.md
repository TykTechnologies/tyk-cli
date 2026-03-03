# Acceptance Review: OAS Versioning

## Coverage Analysis

### Story-to-Scenario Mapping

| Story | AC Count | Scenarios | All ACs Covered |
|-------|----------|-----------|-----------------|
| US-01 | 4 | 5 (WS-01, S-01.2-01.5) | Yes |
| US-02 | 6 | 7 (S-02.1-02.7) | Yes |
| US-03 | 5 | 5 (S-03.1-03.5) | Yes |
| US-04 | 5 | 6 (S-04.1-04.6) | Yes |
| US-05 | 4 | 4 (S-05.1-05.4) | Yes |

### Error Path Coverage

Total scenarios: 27. Error/edge scenarios: 15. Ratio: 56%.

Error paths covered per exit code:
- Exit 1 (auth failure): S-01.5
- Exit 2 (validation): S-02.5
- Exit 3 (not found): S-01.4, S-02.6, S-03.4, S-03.5, S-05.1, S-05.3
- Exit 4 (conflict): S-02.4, S-04.6, S-05.2

### Gaps and Risks

1. **No explicit test for `--version-name` on `tyk api get`** (FR-2). This is listed in requirements but has no dedicated user story. The error helper tests in US-05 (S-05.1, S-05.4) partially cover the version-not-found path for get. If a US is added later, add scenarios.

2. **Batch apply error propagation with `--fail-fast`**: The existing apply tests cover fail-fast behavior. Version-specific fail-fast is implicitly covered by S-04.6 but a dedicated scenario could be added if the crafter finds edge cases.

3. **Concurrent version creation race condition**: Two users creating the same version simultaneously. The pre-check pattern (list then create) has a TOCTOU window. This is accepted per ADR-001 and not tested at the acceptance level.

## Mandate Compliance Evidence

### CM-A: Driving Port Usage

All tests invoke commands through:
- `cmd.RunE(cmd, []string{})` -- direct cobra command execution (matches `executePolicyListCmd` pattern)
- `root.Execute()` -- full command tree execution (matches `executeTykApply` pattern)

No tests call `client.ListOASAPIVersions()` or other internal methods directly. The driving port is the CLI command interface.

### CM-B: Business Language in Scenarios

Zero technical terms in scenario descriptions. All scenarios use:
- "user runs" (not "HTTP request")
- "stderr contains" / "stdout contains" (observable output, not implementation)
- "exit code is" (user-observable behavior)
- "Dashboard receives" (used only to verify mock interactions, framed as the external system)

Terms intentionally absent from scenarios: HTTP, REST, endpoint, handler, struct, interface, goroutine, context, middleware.

### CM-C: Walking Skeleton and Scenario Counts

- Walking skeletons: 3 (WS-01 list, WS-02 create, WS-03 switch-default)
- Focused scenarios: 24
- Total: 27

## Implementation Notes for Software Crafter

### Test Infrastructure to Create

1. **`internal/cli/api_versions_test.go`**: New file in `cli` package. Follow `policy_test.go` patterns:
   - `createVersionConfig(serverURL)` helper returning `*types.Config`
   - `executeVersionsListCmd()`, `executeVersionsCreateCmd()`, `executeVersionsSwitchDefaultCmd()` helpers
   - Mock server handlers for `/api/apis/oas/{id}/versions` (GET) and `/api/apis/oas` (POST)
   - Capture stdout/stderr with `os.Pipe()` or `cmd.SetOut()`/`cmd.SetErr()`

2. **`test/apply_version_integration_test.go`**: New file in `test` package. Follow `apply_integration_test.go` patterns:
   - Extend `mockOpts` or create version-specific mock options
   - Use `executeTykApply()` helper
   - `sampleVersionFile(apiID, baseAPIID, versionName)` fixture helper

3. **OAS fixture files**: Create in `t.TempDir()` per test. Use `writeFixtureFile()` pattern from existing tests.

### Mock Server Endpoints Needed

| Endpoint | Method | Used By |
|----------|--------|---------|
| `/api/apis/oas/{id}/versions` | GET | US-01, US-02, US-03 |
| `/api/apis/oas` | POST | US-02, US-04 |
| `/api/apis/oas/{id}` | GET | US-02, US-05 (detail fetch) |
| `/api/apis/oas/{id}` | PATCH | US-03 |

### Skip Pattern

All tests except WS-01 should start with `t.Skip("not yet implemented -- enable after previous scenario passes")`. This enforces one-at-a-time TDD progression.
