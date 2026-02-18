# Component Boundaries: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DESIGN
**Date**: 2026-02-18

---

## 1. Package Structure

```
tyk-cli/
  internal/
    cli/
      policy.go          [NEW]  Cobra commands + RunE functions
      root.go            [MOD]  1-line addition: rootCmd.AddCommand(NewPolicyCommand())
    client/
      client.go          [---]  Existing HTTP infrastructure (reused)
      policy.go          [NEW]  Policy CRUD methods on Client struct
    policy/              [NEW PACKAGE]
      selector.go        [NEW]  Selector resolution logic
      duration.go        [NEW]  Duration string parsing
      validate.go        [NEW]  Schema validation
      convert.go         [NEW]  CLI <-> wire format conversion
    filehandler/
      filehandler.go     [---]  Reused as-is
    oas/
      transform.go       [---]  Reused as-is
  pkg/
    types/
      policy.go          [NEW]  Policy type definitions
      api.go             [---]  Reused as-is
      config.go          [---]  Reused as-is
```

**New files**: 6
**Modified files**: 1 (root.go, 1 line)
**Reused unchanged**: 7+

---

## 2. Interface Boundaries Between Layers

### CLI Layer -> Policy Logic Layer

The CLI layer (`internal/cli/policy.go`) calls into `internal/policy/` for all business logic. The CLI layer is responsible for:
- Cobra command setup (flags, args, help text)
- Extracting config and output format from context
- Creating the client
- Calling policy logic functions
- Formatting output (human vs JSON)
- Returning `ExitError` with correct codes

The CLI layer does NOT:
- Implement selector resolution logic
- Parse durations
- Validate schema structure
- Convert between CLI and wire formats

### Policy Logic Layer -> Client Layer

The `internal/policy/` package depends on:
- `pkg/types/` for type definitions
- `internal/client/` for API list fetching (selector resolution needs `ListAPIsDashboard`)

The policy logic layer receives the client as a parameter (dependency injection via function args). It does not construct clients itself.

### Client Layer -> Types Layer

`internal/client/policy.go` adds methods to the existing `Client` struct. These methods:
- Use `doRequest` and `handleResponse` from `client.go`
- Accept and return types from `pkg/types/policy.go`
- Follow the exact same patterns as `GetOASAPI`, `CreateOASAPI`, etc.

### Types Layer (Shared)

`pkg/types/policy.go` is imported by all layers. It contains:
- CLI schema types (what users write in YAML)
- Wire types (what the Dashboard API expects/returns)
- No methods with external dependencies (pure data types)

---

## 3. Shared Infrastructure with API Module

### Shared (Used by Both api.go and policy.go)

| Component | Package | Used How |
|---|---|---|
| `GetConfigFromContext` | `internal/cli/context.go` | Both command sets get config same way |
| `GetOutputFormatFromContext` | `internal/cli/context.go` | Both check `--json` flag same way |
| `ExitError{Code, Message}` | `internal/cli/errors.go` | Both use same exit code pattern |
| `client.NewClient(config)` | `internal/client/client.go` | Both create client same way |
| `client.doRequest` | `internal/client/client.go` | Both use same HTTP infrastructure |
| `client.handleResponse` | `internal/client/client.go` | Both use same error handling |
| `client.ListAPIsDashboard` | `internal/client/client.go` | Policy selector resolution + API list display |
| `filehandler.LoadFile` | `internal/filehandler/` | Both load YAML/JSON files same way |
| `types.Config` | `pkg/types/config.go` | Both use same config struct |
| `types.ErrorResponse` | `pkg/types/api.go` | Both handle same API errors |
| `types.OutputFormat` | `pkg/types/config.go` | Both check same output format |
| `truncateWithEllipsis` | `internal/cli/api.go` | Policy list table display |
| `computeTableLayout` | `internal/cli/api.go` | Policy list table sizing |

### Policy-Specific (Not Shared)

| Component | Package | Why Not Shared |
|---|---|---|
| Selector resolution | `internal/policy/selector.go` | Unique to policies; APIs use direct IDs |
| Duration parsing | `internal/policy/duration.go` | API module has no duration concept |
| Schema validation | `internal/policy/validate.go` | Policy YAML schema differs from OAS |
| Wire conversion | `internal/policy/convert.go` | Policy field mapping differs from OAS |
| Policy types | `pkg/types/policy.go` | Different resource model |

### Note on Table Display Helpers

`truncateWithEllipsis` and `computeTableLayout` are currently unexported functions in `api.go`. For policy list to reuse them, one of:
- (a) The crafter exports them (capitalize first letter) in `api.go` -- minimal change
- (b) The crafter duplicates them in `policy.go` -- acceptable for 2 small functions
- (c) The crafter extracts them to a shared `internal/cli/display.go` -- cleanest but more files

The crafter decides which approach during implementation. The architecture supports all three.

---

## 4. Testing Strategy Per Layer

### pkg/types/policy.go -- Type Tests

- **Scope**: Serialization round-trips (YAML marshal/unmarshal, JSON marshal/unmarshal)
- **Pattern**: Table-driven tests with `testify/assert`
- **Dependencies**: None (pure types)
- **Example assertions**: PolicyFile marshals to expected YAML; DashboardPolicy unmarshals from example JSON

### internal/policy/duration.go -- Unit Tests

- **Scope**: `ParseDuration` and `FormatDuration` with edge cases
- **Pattern**: Table-driven, extensive edge cases (zero, negative, invalid suffix, overflow)
- **Dependencies**: None (pure functions)
- **Example assertions**: `ParseDuration("30d") == 2592000`, `ParseDuration("abc") returns error`

### internal/policy/validate.go -- Unit Tests

- **Scope**: Validation error collection for invalid schemas
- **Pattern**: Table-driven with invalid YAML payloads
- **Dependencies**: `pkg/types/` only
- **Example assertions**: Missing metadata.id produces `ValidationError{Field: "metadata.id"}`

### internal/policy/selector.go -- Unit Tests

- **Scope**: Resolution logic with mock API lists
- **Pattern**: Table-driven with crafted API lists testing each selector type
- **Dependencies**: `pkg/types/` only (API list is a plain slice, no client needed)
- **Example assertions**: `name: "users-api"` resolves to `"a1b2c3d4e5f6"` from mock list

### internal/policy/convert.go -- Unit Tests

- **Scope**: Bidirectional conversion (CLI -> wire, wire -> CLI)
- **Pattern**: Round-trip tests: convert CLI -> wire -> CLI, verify equivalence
- **Dependencies**: `pkg/types/` only
- **Example assertions**: `rateLimit.requests: 1000` converts to `rate: 1000` in wire format

### internal/client/policy.go -- Integration-Style Unit Tests

- **Scope**: HTTP request formation, response parsing, error handling
- **Pattern**: `httptest.NewServer` mock (same pattern as `client_test.go`)
- **Dependencies**: `net/http/httptest`, `pkg/types/`
- **Example assertions**: `ListPolicies` sends `GET /api/portal/policies?p=1` with correct auth header

### internal/cli/policy.go -- Command Tests

- **Scope**: End-to-end command execution with mock server
- **Pattern**: Same as `api_get_test.go` -- create mock server, build command, capture stdout/stderr
- **Dependencies**: `httptest`, full dependency chain
- **Example assertions**: `tyk policy list` with mock returns table output; exit code 0

### Test File Convention (Matching Existing)

| Source File | Test File |
|---|---|
| `internal/policy/duration.go` | `internal/policy/duration_test.go` |
| `internal/policy/selector.go` | `internal/policy/selector_test.go` |
| `internal/policy/validate.go` | `internal/policy/validate_test.go` |
| `internal/policy/convert.go` | `internal/policy/convert_test.go` |
| `internal/client/policy.go` | `internal/client/policy_test.go` |
| `internal/cli/policy.go` | `internal/cli/policy_test.go` |
| `pkg/types/policy.go` | `pkg/types/policy_test.go` |

---

## 5. Dependency Graph (Build Order)

```
pkg/types/policy.go           -- no internal deps (build first)
        |
        v
internal/policy/duration.go   -- depends on: types
internal/policy/validate.go   -- depends on: types
internal/policy/selector.go   -- depends on: types
internal/policy/convert.go    -- depends on: types, duration
        |
        v
internal/client/policy.go     -- depends on: types, client.go
        |
        v
internal/cli/policy.go        -- depends on: types, policy/*, client, filehandler, cli/context, cli/errors
        |
        v
internal/cli/root.go          -- adds NewPolicyCommand() (1 line mod)
```

No circular dependencies. Each layer depends only on the layer below it.
