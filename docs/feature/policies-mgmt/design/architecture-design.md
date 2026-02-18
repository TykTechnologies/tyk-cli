# Architecture Design: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DESIGN
**Date**: 2026-02-18
**Status**: Draft

---

## 1. Codebase Analysis (Evidence of Existing Patterns)

### Existing Layered Architecture

```
cmd/main.go                  Entry point, exit code handling
internal/cli/                Cobra commands (api.go, config.go, init.go, root.go)
internal/cli/context.go      Context helpers (config, output format)
internal/cli/errors.go       ExitError type
internal/client/client.go    HTTP client (CRUD against Dashboard)
internal/config/             Config manager (TOML loading)
internal/filehandler/        File loading (YAML/JSON auto-detect)
internal/oas/transform.go    OAS helpers (extension extraction, listen path gen)
pkg/types/api.go             API DTOs (OASAPI, APIResponse, ErrorResponse)
pkg/types/config.go          Config types, exit codes, output format constants
```

### Patterns Extracted from `api.go` (THE reference)

| Pattern | Location | Reuse Strategy |
|---|---|---|
| Command tree: `NewXCommand()` returning `*cobra.Command` | `api.go:154-172` | Mirror for `NewPolicyCommand()` |
| `RunE` functions with `GetConfigFromContext` + `client.NewClient` | `api.go:391-410` | Identical pattern |
| Output split: stderr=human, stdout=data | `api.go:524-531` | Identical pattern |
| JSON output via `GetOutputFormatFromContext` | `api.go:433-446` | Identical pattern |
| `ExitError{Code, Message}` for typed exit codes | `errors.go:4-11` | Reuse directly |
| `filehandler.LoadFile` for YAML/JSON loading | `api.go:971-976` | Reuse directly |
| Delete confirmation prompt | `api.go:1205-1213` | Mirror pattern |
| `ListAPIsDashboard` for API inventory | `client.go:287-363` | Reuse for selector resolution |
| `doRequest`/`handleResponse` HTTP helpers | `client.go:74-159` | Reuse for policy endpoints |
| Table display with `computeTableLayout` | `api.go:47-117` | Adapt for policy columns |

### Reuse vs New Assessment

| Component | Verdict | Rationale |
|---|---|---|
| `client.NewClient` | REUSE | Same Dashboard, same auth, same headers |
| `client.doRequest`/`handleResponse` | REUSE | Policy endpoints follow same REST pattern |
| `filehandler.LoadFile` | REUSE | Policy YAML loaded same way as OAS YAML |
| `cli.ExitError` | REUSE | Same exit code semantics |
| `cli.GetConfigFromContext` | REUSE | Identical config flow |
| `cli.GetOutputFormatFromContext` | REUSE | Identical output format |
| `computeTableLayout` | REUSE | Adapt column widths for policy list |
| Selector resolution | NEW | No existing resolver; `ListAPIsDashboard` exists but needs wrapper |
| Duration parser | NEW | No existing duration handling in codebase |
| Policy types | NEW | No policy DTOs exist yet |
| Policy client methods | NEW | New CRUD endpoints, but follows `client.go` pattern exactly |
| CLI-to-wire conversion | NEW | Field mapping unique to policies |

---

## 2. Component Architecture

### Layer Diagram

```
+---------------------------------------------------------------------+
|  cmd/main.go                                                         |
|  (entry point, exit code handling -- EXISTING, no changes)           |
+---------------------------------------------------------------------+
         |
         v
+---------------------------------------------------------------------+
|  internal/cli/                                                       |
|  root.go  -- adds NewPolicyCommand() to rootCmd  [MODIFY: 1 line]   |
|  policy.go -- cobra command tree + RunE functions [NEW]              |
|  context.go, errors.go                            [REUSE as-is]     |
|  api.go                                           [REUSE as-is]     |
+---------------------------------------------------------------------+
         |
         v
+---------------------------------------------------------------------+
|  internal/policy/                                  [NEW PACKAGE]     |
|  selector.go  -- resolve name/listenPath/id/tags to API IDs         |
|  duration.go  -- parse "30d"/"1m"/"60" to seconds                   |
|  validate.go  -- schema validation for policy YAML                  |
|  convert.go   -- CLI schema <-> Dashboard wire format conversion     |
+---------------------------------------------------------------------+
         |
         v
+---------------------------------------------------------------------+
|  internal/client/                                                    |
|  client.go  -- existing HTTP infrastructure       [REUSE]           |
|  policy.go  -- ListPolicies, GetPolicy,           [NEW]             |
|                CreatePolicy, UpdatePolicy,                           |
|                DeletePolicy                                          |
+---------------------------------------------------------------------+
         |
         v
+---------------------------------------------------------------------+
|  pkg/types/                                                          |
|  policy.go  -- CLI schema types, wire types,       [NEW]            |
|                conversion type definitions                           |
|  api.go     -- OASAPI, ErrorResponse, etc.         [REUSE as-is]   |
|  config.go  -- Config, ExitCode, OutputFormat      [REUSE as-is]   |
+---------------------------------------------------------------------+
         |
         v
+---------------------------------------------------------------------+
|  internal/filehandler/                                               |
|  filehandler.go -- LoadFile, YAML/JSON auto-detect [REUSE as-is]   |
+---------------------------------------------------------------------+
```

### Files to Create

| File | Responsibility |
|---|---|
| `internal/cli/policy.go` | Cobra command tree: `NewPolicyCommand()` with `list`, `get`, `apply`, `delete`, `init` subcommands. RunE functions for each. |
| `internal/client/policy.go` | Dashboard policy CRUD: `ListPolicies`, `GetPolicy`, `CreatePolicy`, `UpdatePolicy`, `DeletePolicy`. Follows `doRequest`/`handleResponse` pattern from `client.go`. |
| `pkg/types/policy.go` | CLI schema types (`PolicyFile`, `PolicyMetadata`, `PolicySpec`, `AccessEntry`, `Selector`), wire types (`DashboardPolicy`), and conversion type definitions. |
| `internal/policy/selector.go` | Selector resolution: takes `[]AccessEntry` + API list, returns resolved `map[selector]->apiID(s)` or structured errors (ambiguous, not found with suggestions). |
| `internal/policy/duration.go` | Duration parser: `ParseDuration("30d") -> 2592000`, `FormatDuration(2592000) -> "30d"`. |
| `internal/policy/validate.go` | Schema validation: required fields, type checks, duration format, selector format. Returns `[]ValidationError` with field paths. |
| `internal/policy/convert.go` | Bidirectional conversion: CLI schema -> Dashboard wire format (for apply), Dashboard wire -> CLI schema (for get). |

### Files to Modify

| File | Change |
|---|---|
| `internal/cli/root.go` | Add `rootCmd.AddCommand(NewPolicyCommand())` -- 1 line |

---

## 3. Command Flow Architecture

### `tyk policy list`

```
1. GetConfigFromContext -> client.NewClient
2. client.ListPolicies(ctx, page)
   -> GET /api/policies?p={page}
   -> handleResponse -> []DashboardPolicy
3. Format output:
   - Human: table with ID, Name, APIs count, Tags columns to stdout; header to stderr
   - JSON: {page, count, policies} to stdout
```

### `tyk policy get <id>`

```
1. GetConfigFromContext -> client.NewClient
2. client.GetPolicy(ctx, policyID)
   -> GET /api/policies/{policyId}
   -> handleResponse -> DashboardPolicy
3. convert.WireToCLI(dashboardPolicy, apiList)
   -> Reverse-resolve API IDs to names (best-effort via ListAPIsDashboard)
   -> Convert seconds to duration strings
   -> Build PolicyFile struct
4. Format output:
   - Human: summary to stderr, CLI schema YAML to stdout
   - JSON: CLI schema JSON to stdout
5. Not found -> ExitError{Code: 3}
```

### `tyk policy apply -f <file>`

```
1. Load file: filehandler.LoadFile or stdin
2. Parse into PolicyFile struct (YAML unmarshal)
3. validate.ValidatePolicy(policyFile) -> []ValidationError
   - Required fields: metadata.id, metadata.name
   - Duration format validation
   - Selector format validation (exactly one of id/name/listenPath/tags per access entry)
4. Resolve selectors:
   a. client.ListAPIsDashboard(ctx, allPages) -> complete API inventory
   b. selector.ResolveAll(policyFile.Spec.Access, apiList)
      -> For each entry: match selector to API(s)
      -> Fail-fast on any resolution error (before mutations)
5. Convert durations: duration.ParseDuration for per, period, keyTTL
6. convert.CLIToWire(policyFile, resolvedAPIs) -> DashboardPolicy
7. Upsert:
   a. client.GetPolicy(ctx, metadata.id) -- check existence
   b. If not found: client.CreatePolicy(ctx, wirePolicy)
   c. If found: client.UpdatePolicy(ctx, id, wirePolicy)
8. Format output:
   - Human: resolution log + success message to stderr
   - JSON: {policy_id, operation, api_count} to stdout
```

### `tyk policy delete <id>`

```
1. GetConfigFromContext -> client.NewClient
2. client.GetPolicy(ctx, policyID) -- verify exists, get name for prompt
3. Confirmation prompt (unless --yes)
4. client.DeletePolicy(ctx, policyID)
   -> DELETE /api/policies/{policyId}
5. Format output (mirrors api delete exactly)
```

### `tyk policy init`

```
1. Prompt for policy ID and name (no Dashboard connectivity needed)
2. Generate scaffold PolicyFile struct with defaults
3. Marshal to YAML with comments
4. Write to policies/{id}.yaml (with overwrite check)
5. Output: "Scaffolded: policies/{id}.yaml" to stderr
```

---

## 4. Selector Resolution Architecture

### Resolution Strategy

Selectors resolve API references in policy YAML to Dashboard API IDs. This runs entirely before any server mutation (fail-fast).

```
Input: []AccessEntry from YAML, each with exactly one selector field set
Dependencies: Complete API list from ListAPIsDashboard (all pages)

For each AccessEntry:
  switch selector type:
    case "id":
      -> Direct lookup in API list
      -> Must match exactly 1 API, else error (not found)
    case "name":
      -> Filter API list by Name == selector value
      -> Must match exactly 1 API
      -> 0 matches: error with fuzzy suggestions (Levenshtein distance)
      -> >1 matches: error with candidate list + disambiguation guidance
    case "listenPath":
      -> Filter API list by ListenPath == selector value
      -> Same uniqueness rules as "name"
    case "tags":
      -> Filter API list by APIs containing ALL specified tags (AND logic)
      -> Must match >= 1 API
      -> 0 matches: error listing available tags
      -> Multiple matches: valid (tags expand to many APIs)
```

### Fuzzy Suggestion Strategy

When a name or listenPath selector matches zero APIs, provide "Did you mean?" suggestions using string distance. The algorithm computes edit distance between the selector value and all API names/paths, returning the top 3 closest matches. This can be implemented with the standard Levenshtein algorithm (no external library needed for this small dataset -- the API list is typically under 1000 items).

### API List Caching During Apply

A single `apply` invocation may need to resolve multiple selectors. The API list is fetched once and reused for all resolutions within the same command. If the Dashboard paginates at 10 items/page and there are 100 APIs, this means 10 sequential fetches. For Phase 1, this is acceptable. If performance becomes an issue, a batch endpoint or parallel fetching can be added.

---

## 5. Duration Parsing Architecture

### Supported Formats

| Input | Output (seconds) | Notes |
|---|---|---|
| `60` | 60 | Plain integer passthrough |
| `60s` | 60 | Seconds suffix |
| `1m` | 60 | Minutes |
| `1h` | 3600 | Hours |
| `30d` | 2592000 | Days (24h each) |
| `0` | 0 | Special: no expiry |

### Rules

- Integer-only input is treated as seconds
- Single suffix character: `s`, `m`, `h`, `d`
- No fractional values (e.g., `1.5h` is rejected)
- No mixed units (e.g., `1h30m` is rejected)
- Negative values are rejected
- `FormatDuration` reverses: picks largest clean unit (e.g., 86400 -> "1d", 90 -> "90s")

### Implementation Note

No external library needed. This is a simple regex + switch on suffix. Approximately 30 lines of Go. The parser returns `(int64, error)` -- the crafter decides the internal structure.

---

## 6. Schema Validation Approach

### Validation Order (Fail-Fast Pipeline)

1. **YAML parse** -- file is valid YAML (handled by `filehandler.LoadFile`)
2. **Schema structure** -- `apiVersion`, `kind`, `metadata`, `spec` present
3. **Required fields** -- `metadata.id`, `metadata.name` non-empty
4. **Type checks** -- `rateLimit.requests` is integer, `access` is list, etc.
5. **Duration parsing** -- `per`, `period`, `keyTTL` are valid duration strings
6. **Selector format** -- each access entry has exactly one of `id`/`name`/`listenPath`/`tags`
7. **Selector resolution** -- resolved against live Dashboard data (separate step, after local validation)

### Error Reporting

Validation returns a list of errors, each with:
- Field path (e.g., `spec.access[0].name`)
- Error message (e.g., "no API found for name 'inventori-api'")
- Error kind (schema, resolution, duration)

All local validation errors (steps 2-6) are collected and reported together. Selector resolution errors (step 7) are separate since they require network access.

---

## 7. Error Handling Strategy

### Exit Codes (Matching Existing Patterns)

| Code | Meaning | Policy Usage |
|---|---|---|
| 0 | Success | All successful operations including empty list |
| 1 | General failure | Network errors, unexpected server responses |
| 2 | Bad arguments | Missing file, invalid flag combo, schema validation, ambiguous/missing selectors |
| 3 | Not found | `get`/`delete` for non-existent policy ID |
| 4 | Conflict | Reserved (not expected for Phase 1) |

### Error Pattern (Matching `api.go`)

```
return &ExitError{Code: 3, Message: fmt.Sprintf("policy '%s' not found", policyID)}
return &ExitError{Code: 2, Message: "selector ambiguous..."}
return fmt.Errorf("failed to create client: %w", err)  // -> exit 1
```

### Output Convention (Matching Existing)

- Human-readable messages: stderr
- Data (YAML, JSON, tables): stdout
- Error messages: stderr (via `cmd/main.go` error handler)
- Colored output: `fatih/color` on stderr only

---

## 8. Dashboard API Endpoints

| Operation | Method | Path | Notes |
|---|---|---|---|
| List | GET | `/api/portal/policies?p={page}` | Paginated; returns policy array |
| Get | GET | `/api/portal/policies/{policyId}` | Returns single policy JSON |
| Create | POST | `/api/portal/policies` | Returns created policy with ID |
| Update | PUT | `/api/portal/policies/{policyId}` | Returns updated policy |
| Delete | DELETE | `/api/portal/policies/{policyId}` | Returns status |

Note: The exact endpoint prefix (`/api/portal/policies` vs `/api/policies`) must be confirmed against the target Dashboard version. The architecture supports either -- it is a single constant in `internal/client/policy.go`.

---

## 9. C4 Diagrams

### Context Diagram

```
+-------------------+          +---------------------+
|                   |  REST    |                     |
|  Platform         |--------->|  Tyk Dashboard      |
|  Engineer (Ravi)  |  HTTP    |  REST API           |
|                   |          |                     |
+-------------------+          +---------------------+
        |                              ^
        | CLI commands                 | Policies + APIs
        v                              | stored here
+-------------------+                  |
|                   |------------------+
|  tyk CLI          |
|  (Go binary)      |
|                   |---------> Policy YAML files
+-------------------+           (local disk, Git)
```

### Container Diagram

```
+------------------------------------------------------------------+
|  tyk CLI Binary                                                   |
|                                                                   |
|  +-------------------+  +-------------------+  +--------------+  |
|  |  cmd/main.go      |  |  internal/cli/    |  | internal/    |  |
|  |  Entry point      |->|  Command layer    |->| policy/      |  |
|  |  Exit codes       |  |  api.go           |  | selector.go  |  |
|  +-------------------+  |  policy.go [NEW]  |  | duration.go  |  |
|                          |  config.go        |  | validate.go  |  |
|                          |  root.go          |  | convert.go   |  |
|                          +-------------------+  +--------------+  |
|                                    |                   |          |
|                                    v                   v          |
|  +-------------------+  +-------------------+  +--------------+  |
|  |  internal/        |  |  internal/client/  |  | pkg/types/   |  |
|  |  filehandler/     |  |  client.go         |  | api.go       |  |
|  |  YAML/JSON load   |  |  policy.go [NEW]   |  | policy.go    |  |
|  +-------------------+  +-------------------+  | [NEW]        |  |
|                                    |            | config.go    |  |
|                                    |            +--------------+  |
+------------------------------------------------------------------+
                                     |
                                     | HTTP REST
                                     v
                          +-------------------+
                          | Tyk Dashboard API |
                          | /api/portal/      |
                          | policies          |
                          +-------------------+
```

### Component Diagram (Policy Module Internals)

```
+----------------------------------------------------------------------+
|  internal/policy/                                                     |
|                                                                       |
|  +-------------------+                                                |
|  |  validate.go      |  Schema validation pipeline                    |
|  |                   |  Input: raw map[string]interface{}              |
|  |                   |  Output: PolicyFile or []ValidationError       |
|  +-------------------+                                                |
|           |                                                           |
|           v                                                           |
|  +-------------------+      +-------------------+                     |
|  |  selector.go      |<---->|  (client.          |                    |
|  |                   |      |  ListAPIsDashboard)|                    |
|  |  ResolveAll()     |      +-------------------+                     |
|  |  Input: []Access  |                                                |
|  |  Output: resolved |                                                |
|  |    API IDs or err |                                                |
|  +-------------------+                                                |
|           |                                                           |
|           v                                                           |
|  +-------------------+                                                |
|  |  duration.go      |  ParseDuration("30d") -> 2592000               |
|  |                   |  FormatDuration(2592000) -> "30d"              |
|  +-------------------+                                                |
|           |                                                           |
|           v                                                           |
|  +-------------------+                                                |
|  |  convert.go       |  CLIToWire: PolicyFile -> DashboardPolicy      |
|  |                   |  WireToCLI: DashboardPolicy -> PolicyFile      |
|  +-------------------+                                                |
|                                                                       |
+----------------------------------------------------------------------+
```

---

## 10. ADRs

### ADR-001: Policy Module as Sibling Package to OAS

**Status**: Accepted

**Context**: Policies need selector resolution, duration parsing, and schema conversion logic that does not exist in the OAS module. The question is whether to extend `internal/oas/` or create `internal/policy/`.

**Decision**: Create `internal/policy/` as a new package parallel to `internal/oas/`.

**Alternatives Considered**:
- Extend `internal/oas/`: Rejected -- policy logic (selectors, durations, wire conversion) is unrelated to OAS transformation. Mixing concerns violates SRP and creates a confusing package boundary.
- Put everything in `internal/cli/policy.go`: Rejected -- the `api.go` file is already 1594 lines with mixed concerns. Separating business logic into `internal/policy/` keeps CLI commands thin.

**Consequences**:
- Positive: Clear separation of concerns; policy logic testable without cobra dependency
- Positive: Matches team's layered architecture convention
- Negative: One additional package to navigate

### ADR-002: No New External Dependencies for Phase 1

**Status**: Accepted

**Context**: Policies need duration parsing and fuzzy string matching. Should we add libraries?

**Decision**: Implement both with standard library only. No new dependencies for Phase 1.

**Alternatives Considered**:
- `github.com/agnivade/levenshtein` (MIT): Well-maintained, but edit distance on a list of <1000 items is trivial to implement in ~20 lines. Adding a dependency for 20 lines is overhead.
- `github.com/lithammer/fuzzysearch` (MIT): More features than needed. We only need ranked suggestions from a small list.
- Go `time.ParseDuration`: Built-in Go duration parser supports `h`, `m`, `s`, `ms` but NOT `d` (days). Policy durations require days. A thin wrapper would still need custom logic for `d`, and the stdlib parser's `ns`/`us`/`ms` suffixes are not relevant to policy durations.

**Consequences**:
- Positive: Zero dependency bloat; `go.mod` unchanged
- Positive: Full control over duration format (exactly `s/m/h/d`)
- Negative: ~50 lines of manual implementation vs library call
- Revisit: If Phase 2 `bind/unbind` needs richer fuzzy matching, reconsider

### ADR-003: Selector Resolution Fetches All APIs Once Per Apply

**Status**: Accepted

**Context**: Selector resolution needs the complete API list to match names, listen paths, and tags. The Dashboard paginates at 10 APIs per page. Should we fetch lazily or eagerly?

**Decision**: Fetch all pages eagerly at the start of apply, cache in memory for the duration of the command.

**Alternatives Considered**:
- Lazy fetch per selector: Rejected -- multiple selectors would make redundant API calls, and tag selectors need the full list regardless.
- Add a "list all" endpoint to the client: Same result (fetch all pages), just wrapped differently. Could be added if useful for other commands.

**Consequences**:
- Positive: Simple, predictable, correct (tag selectors always need full list)
- Positive: All resolution errors reported together (good UX)
- Negative: Slow for very large API counts (>500 APIs, >50 page fetches). Acceptable for Phase 1; optimize if measured.

### ADR-004: Policy Client Methods on Existing Client Struct

**Status**: Accepted

**Context**: Where should policy CRUD methods live -- on the existing `Client` struct or a new `PolicyClient`?

**Decision**: Add methods to the existing `Client` struct in a new file `internal/client/policy.go`.

**Alternatives Considered**:
- New `PolicyClient` struct: Rejected -- would duplicate HTTP infrastructure (`doRequest`, `handleResponse`, auth headers). The existing `Client` already has `ListAPIsDashboard` which policy resolution depends on.
- Separate package `internal/client/policy/`: Rejected -- unnecessary package nesting for methods on the same struct.

**Consequences**:
- Positive: Reuses all HTTP infrastructure; no duplication
- Positive: Policy resolution can call `c.ListAPIsDashboard()` directly
- Negative: `Client` struct grows (but still single-responsibility: "Dashboard API client")

---

## 11. Walking Skeleton Recommendation

**Start with**: `tyk policy list` + `tyk policy apply -f` (name selector only)

**Rationale**: These two commands validate the complete architecture vertically:
- `list` validates: client CRUD, wire types, output formatting
- `apply` validates: file loading, schema validation, selector resolution, duration parsing, CLI-to-wire conversion, upsert logic

**Phase 1 implementation order**:
1. Types (`pkg/types/policy.go`) -- foundation for all other work
2. Client CRUD (`internal/client/policy.go`) -- enables list/get/apply/delete
3. Duration parser (`internal/policy/duration.go`) -- needed by apply
4. Schema validation (`internal/policy/validate.go`) -- needed by apply
5. Selector resolution (`internal/policy/selector.go`) -- needed by apply
6. Wire conversion (`internal/policy/convert.go`) -- needed by apply and get
7. CLI commands (`internal/cli/policy.go`) -- wires everything together
8. Root registration (`internal/cli/root.go` -- 1 line change)

---

## 12. Phase 2 Integration Points (Future)

Phase 2 commands (`who-uses`, `bind`, `unbind`) build on Phase 1 infrastructure:

- `who-uses`: Reuses `ListPolicies` + `ListAPIsDashboard` + selector resolution
- `bind/unbind`: Reuses `GetPolicy` + `UpdatePolicy` + selector resolution

No architectural changes needed for Phase 2 -- only new CLI commands and minor client method additions.
