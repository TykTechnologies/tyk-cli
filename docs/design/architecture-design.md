# Architecture Design: OAS Versioning Commands

## System Context

Single Go binary (Cobra CLI) communicating synchronously with Tyk Dashboard REST API. No local state. All version data lives on the Dashboard.

## Existing System Analysis

### Reusable Components

| Component | Location | Reuse |
|---|---|---|
| `client.ListOASAPIVersions()` | `internal/client/client.go:428` | Direct use for US-01, US-03, US-05 |
| `client.SwitchDefaultVersion()` | `internal/client/client.go:445` | Direct use for US-03 |
| `client.GetOASAPI()` | `internal/client/client.go:175` | Direct use for US-05 (API name lookup) |
| `types.CreateOASAPIRequest` | `pkg/types/api.go:49` | Direct use for US-02 (has BaseAPIID, NewVersionName, SetDefault) |
| `types.VersionListResponse` | `pkg/types/api.go:67` | Direct use for US-01 |
| `ExitError` | `internal/cli/errors.go:3` | Direct use for all exit codes |
| `GetConfigFromContext()` | `internal/cli/context.go:22` | Standard pattern |
| `GetOutputFormatFromContext()` | `internal/cli/context.go:35` | Standard pattern |
| `extractVersionFromOAS()` | `internal/cli/api.go:841` | Reuse for version name extraction |
| `oas.HasTykExtensions()` | `internal/oas/transform.go:13` | Reuse for classification |
| `oas.ExtractAPIIDFromTykExtensions()` | `internal/oas/transform.go:19` | Reuse for ID extraction |
| `classifyContent()` | `internal/cli/apply.go:447` | Extend for version file detection |
| `sortFiles()` | `internal/cli/apply.go:465` | Extend for version ordering |
| `filehandler.LoadFile()` | `internal/filehandler/` | Reuse for OAS file loading |
| Placeholder commands | `internal/cli/api.go:185-216` | Replace in-place |

### What Must Be New

| Component | Justification |
|---|---|
| `client.CreateOASAPIVersion()` | No existing client method posts `CreateOASAPIRequest` with BaseAPIID |
| `oas.ExtractVersioningMetadata()` | No existing function reads `x-tyk-api-gateway.info.versioning` |
| Version error helpers (US-05) | No existing helpers format version-aware errors with hints |

## Component Changes

### Layer 1: `internal/client/client.go` -- New Client Method

**New method: `CreateOASAPIVersion`**

- Accepts `CreateOASAPIRequest` (already defined in types)
- POSTs to `POST /api/apis/oas` with `base_api_id`, `new_version_name`, `set_default` in query params or body
- Returns created API metadata

```
Data flow:
  CreateOASAPIVersion(ctx, req CreateOASAPIRequest)
    -> POST /api/apis/oas  (with base_api_id, new_version_name, set_default)
    -> handleResponse() -> GetOASAPI() to fetch full details
    -> returns (*OASAPI, error)
```

No other new client methods needed. All other operations use existing methods.

### Layer 2: `internal/oas/transform.go` -- Versioning Metadata Extraction

**New function: `ExtractVersioningMetadata`**

Reads from the OAS extension convention for batch apply:

```yaml
x-tyk-api-gateway:
  info:
    versioning:
      base_api_id: "abc123"
      version_name: "v3"
      set_default: false
```

Returns a struct with `BaseAPIID`, `VersionName`, `SetDefault`, `IsVersionFile` fields. Returns `IsVersionFile: false` when no versioning section exists.

### Layer 3: `internal/cli/api.go` -- Command Implementations

Replace three placeholder functions (lines 185-216) with full implementations.

#### `NewAPIVersionsListCommand()` (US-01)

- Flags: `--api-id` (required)
- Flow: `GetConfigFromContext` -> `client.NewClient` -> `client.ListOASAPIVersions(ctx, apiID)`
- Human output (stderr): table of versions with default marker, count summary
- JSON output (stdout): `{"api_id", "api_name", "default", "versions": [...]}`
- Error: 404 -> exit code 3 with API-not-found message (US-05)

#### `NewAPIVersionsCreateCommand()` (US-02)

- Flags: `--api-id` (required), `-f` (required), `--version-name` (required), `--set-default` (optional, default false)
- Flow:
  1. Load OAS file (reuse `filehandler.LoadFile`)
  2. `ListOASAPIVersions` to check for name conflict -> exit 4 if duplicate
  3. `CreateOASAPIVersion` with `CreateOASAPIRequest{BaseAPIID, NewVersionName, SetDefault, OAS}`
  4. Output result; if not set-default, show switch-default hint
- JSON output: `{"action": "created", "api_id", "version_name", "listen_path", "upstream_url", "is_default"}`
- Errors: conflict -> exit 4, invalid OAS -> exit 2, API not found -> exit 3

#### `NewAPIVersionsSwitchDefaultCommand()` (US-03)

- Flags: `--api-id` (required), `--version-name` (required)
- Flow:
  1. `ListOASAPIVersions` to get current default and validate target exists
  2. If target == current default -> no-op message, exit 0
  3. If target not in version list -> exit 3 with available versions (US-05)
  4. `SwitchDefaultVersion(ctx, apiID, versionName)`
  5. Output previous and new default
- JSON output: `{"action": "switched", "api_id", "previous_default", "new_default"}`

### Layer 4: `internal/cli/apply.go` -- Batch Apply Extension (US-04)

#### Classification Extension

- Add `configFileAPIVersion` constant to `configFileType`
- Extend `classifyContent()`: if `HasTykExtensions` AND `ExtractVersioningMetadata().IsVersionFile` -> return `configFileAPIVersion`

#### Ordering Extension

- Extend `typePriority()`: policy=0, api=1, apiVersion=2, unrecognized=3
- This ensures: policies first, then base APIs, then version files

#### Apply Extension

- Extend `applyFileWithOp()` to handle `configFileAPIVersion`
- Extract versioning metadata, build `CreateOASAPIRequest`, POST via `batchApplier.doJSON`
- Dry-run: report "would create version: {name} of {base_api_id}"
- JSON result entry: operation `"version_created"`, add `version_name` field

#### Single File Apply Extension

- Add `--base-api-id` flag to `NewAPIApplyCommand()`
- When `--base-api-id` is set, treat file as version creation (same as US-02 flow)

### Layer 5: `internal/cli/errors.go` or new `internal/cli/version_errors.go` -- Error Helpers (US-05)

Cross-cutting helper functions used by US-01, US-02, US-03, US-04:

- **Version-not-found error**: accepts apiID, requestedVersion, availableVersions -> formats message with available list
- **Version-conflict error**: accepts apiID, versionName -> formats message with apply suggestion
- **API-not-found error**: accepts apiID -> formats message with `tyk api list` suggestion
- **API name enrichment**: optional `GetOASAPI` call to include API name in error messages (best-effort, does not fail if lookup fails)

All helpers return `*ExitError` with appropriate exit codes (3 for not-found, 4 for conflict).

### Layer 6: `pkg/types/api.go` -- No Changes

`CreateOASAPIRequest` already has `BaseAPIID`, `NewVersionName`, `SetDefault` fields. `VersionListResponse` already has `Versions` and `Default`. No new types needed.

### Layer 7: `internal/cli/apply.go` -- jsonResultEntry Extension

Add optional `VersionName` field to `jsonResultEntry`:

```
VersionName string `json:"version_name,omitempty"`
```

## Data Flow Diagrams

### US-01: List Versions

```
User -> tyk api versions list --api-id abc123
  -> cli.runVersionsList()
  -> client.ListOASAPIVersions(ctx, "abc123")
  -> GET /api/apis/oas/abc123/versions
  <- VersionListResponse{Versions: ["v1","v2","v3"], Default: "v1"}
  <- stdout (JSON) or stderr (human table)
```

### US-02: Create Version

```
User -> tyk api versions create --api-id abc123 -f v3.yaml --version-name v3
  -> cli.runVersionsCreate()
  -> filehandler.LoadFile("v3.yaml")
  -> client.ListOASAPIVersions(ctx, "abc123")  [conflict check]
  -> client.CreateOASAPIVersion(ctx, CreateOASAPIRequest{
       OAS: <file content>, BaseAPIID: "abc123",
       NewVersionName: "v3", SetDefault: false
     })
  -> POST /api/apis/oas  (with base_api_id=abc123&new_version_name=v3)
  <- APIResponse{ID: "newid"}
  -> client.GetOASAPI(ctx, "newid", "")  [fetch details]
  <- stdout (JSON) or stderr (human summary + hint)
```

### US-03: Switch Default

```
User -> tyk api versions switch-default --api-id abc123 --version-name v3
  -> cli.runVersionsSwitchDefault()
  -> client.ListOASAPIVersions(ctx, "abc123")  [validate + get current]
  -> client.SwitchDefaultVersion(ctx, "abc123", "v3")
  -> PATCH /api/apis/oas/abc123  {set_default_version: "v3"}
  <- stdout (JSON) or stderr (human: previous=v1, new=v3)
```

### US-04: Batch Apply with Versions

```
User -> tyk apply -f apis/
  -> scanDirectory("apis/")
  -> classifyFile() per file
     -> classifyContent() detects x-tyk-api-gateway.info.versioning -> configFileAPIVersion
  -> sortFiles(): [policies, base APIs, version files]
  -> for each file:
     -> configFileAPIVersion: applyVersionFile()
        -> POST /api/apis/oas (with base_api_id, new_version_name from metadata)
```

## OAS Extension Schema for Versioning

Convention for declaring version relationships in OAS files (batch apply):

```yaml
x-tyk-api-gateway:
  info:
    name: "Petstore"
    versioning:
      base_api_id: "abc123"
      version_name: "v3"
      set_default: false
```

- `base_api_id` (required): Links version to parent API
- `version_name` (required): Version identifier, must be unique within parent
- `set_default` (optional, default false): Whether to set as default on creation

This extension is only consumed by `tyk apply -f` for batch classification. Direct commands use `--api-id` and `--version-name` flags.

## Error Handling Patterns

| Scenario | HTTP Status | Exit Code | Message Pattern |
|---|---|---|---|
| API not found | 404 | 3 | `API not found: {id}. Use 'tyk api list' to see available APIs.` |
| Version not found | 404 (logical) | 3 | `version "{name}" not found for API {id} ({api_name}). Available versions: {v1, v2}.` |
| Version conflict | N/A (pre-check) | 4 | `version "{name}" already exists for API {id}. Use 'tyk api apply -f <file>' to update.` |
| Invalid OAS file | N/A (local) | 2 | `invalid OAS document: {reason}` |
| Auth failure | 401 | 1 | `Authentication failed` |

## Files Changed Summary

| File | Change Type | Scope |
|---|---|---|
| `internal/client/client.go` | Add method | `CreateOASAPIVersion()` (~30 lines) |
| `internal/oas/transform.go` | Add function | `ExtractVersioningMetadata()` (~25 lines) |
| `internal/cli/api.go` | Replace placeholders | 3 command functions (~200 lines replacing ~30) |
| `internal/cli/apply.go` | Extend | classification, ordering, version apply (~60 lines) |
| `internal/cli/version_errors.go` | New file | Error helper functions (~50 lines) |
| `pkg/types/api.go` | No changes | Types already sufficient |

**Total new/changed production files: 5**
**Estimated new lines: ~365**

## Implementation Order

1. **US-05** (version_errors.go) -- cross-cutting helpers, no dependencies
2. **US-01** (versions list) -- foundation for conflict/validation checks
3. **US-03** (switch-default) -- uses existing client method + US-01 for validation
4. **US-02** (create version) -- new client method + US-01 for conflict detection
5. **US-04** (apply integration) -- extends apply with US-02 creation logic

## ADRs

### ADR-001: Pre-check Pattern for Version Validation

**Status**: Accepted

**Context**: Version commands need to validate that versions exist (switch-default) or do not exist (create) before performing mutations. Two approaches: optimistic (try and handle error) or pre-check (list first, then act).

**Decision**: Pre-check via `ListOASAPIVersions` before mutation. One extra API call per command invocation.

**Alternatives Considered**:
- Optimistic mutation with error parsing: Rejected because Dashboard error responses are inconsistent (sometimes 400, sometimes 409, varying message formats). Pre-check is reliable and enables rich error messages with available versions list.
- Client-side caching of version lists: Rejected as unnecessary complexity for a CLI that runs once per invocation.

**Consequences**:
- Positive: Reliable conflict/not-found detection, enables US-05 error messages with available versions
- Negative: One extra HTTP round-trip per version command (~50ms, acceptable for CLI)

### ADR-002: Versioning Extension Convention for Batch Apply

**Status**: Accepted

**Context**: Batch apply (`tyk apply -f dir/`) needs to distinguish version files from standalone API files. Two options: file naming convention or embedded metadata.

**Decision**: Embedded metadata in `x-tyk-api-gateway.info.versioning` section within the OAS file.

**Alternatives Considered**:
- File naming convention (e.g., `*-version-*.yaml`): Rejected because it is fragile, not self-documenting, and cannot carry version parameters (base_api_id, set_default).
- Separate manifest file listing version relationships: Rejected as unnecessary additional file; the OAS file is already the unit of configuration.

**Consequences**:
- Positive: Self-contained, each file declares its own version relationship, no external manifest needed
- Negative: Introduces a Tyk-specific extension sub-section; files are not portable to non-Tyk systems (acceptable since `x-tyk-api-gateway` already makes them Tyk-specific)

### ADR-003: Shared Error Helpers as Separate File

**Status**: Accepted

**Context**: US-05 error helpers are used across US-01, US-02, US-03, US-04. They could live in `api.go` (already large at ~1500 lines) or in a dedicated file.

**Decision**: New file `internal/cli/version_errors.go` in the same `cli` package.

**Alternatives Considered**:
- Add to `api.go`: Rejected because `api.go` is already ~1500 lines and version errors are a distinct concern.
- Add to `errors.go`: Rejected because `errors.go` contains only the generic `ExitError` type; version-specific helpers are a different abstraction level.

**Consequences**:
- Positive: Clear separation, easy to find version error helpers, keeps `api.go` from growing further
- Negative: One additional file (minimal cost)
