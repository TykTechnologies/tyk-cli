---
title: Policy Design
nav_order: 10
---

# Policy System Design

This document explains the architecture, design decisions, and trade-offs behind the `tyk policy` feature. It is aimed at contributors and anyone who needs to understand *why* the system works the way it does.

## The core problem

The Tyk Dashboard API uses a JSON wire format for policies that is hostile to humans. API access is keyed by opaque 24-character hex IDs. Durations are raw seconds. There is no validation client-side — the server silently accepts garbage and produces broken policies. None of this is version-control friendly.

The goal: let users author policies as readable YAML files that they can commit, review, and diff, while the CLI handles all the ugly translation.

## Two-format architecture

The system has two distinct representations for the same policy:

```
┌──────────────────┐          ┌──────────────────┐
│   CLI Schema     │  apply   │   Wire Format    │
│   (PolicyFile)   │ ──────►  │ (DashboardPolicy)│
│   Human-written  │          │   Dashboard API  │
│   YAML files     │  ◄────── │   JSON           │
└──────────────────┘    get   └──────────────────┘
```

**CLI schema** (`pkg/types/PolicyFile`) — what users write:
- Durations as strings: `"30d"`, `"1h"`, `"1m"`
- APIs referenced by name, listen path, or tags
- Flat YAML structure, easy to read and diff

**Wire format** (`pkg/types/DashboardPolicy`) — what the Dashboard expects:
- Durations as integer seconds: `2592000`, `3600`, `60`
- APIs keyed by hex ID in an `access_rights` map
- Nested JSON with fields like `_id`, `quota_renewal_rate`, `key_expires_in`
- `allowed_urls` must be `[]` not `null`, `limit` must be explicit `null` — the Dashboard is picky

The conversion lives in `internal/policy/convert.go`. It is bidirectional:
- `CLIToWire` — used by `apply` to push to the Dashboard. Passes the user's `id` directly into the wire `id` field and leaves `_id` empty (the Dashboard generates it).
- `WireToCLI` — used by `get` to pull from the Dashboard into readable YAML. Uses the wire `id` field directly as the friendly ID, falling back to `_id` for unmanaged policies (those with an empty `id` field).

### Why not just use the wire format?

We considered having users write Dashboard JSON directly (or a thin YAML wrapper over it). Rejected because:

1. **API IDs are meaningless.** `"a1b2c3d4e5f6"` tells you nothing. `name: users-api` tells you everything. IDs also change between environments, so a policy file with hardcoded IDs cannot be promoted from dev to staging.

2. **Durations are error-prone.** Is `2592000` thirty days or some other number? `"30d"` is unambiguous.

3. **The wire format is unstable.** Tyk has added and renamed fields across versions. The CLI schema acts as a stable interface — if the wire format changes, only `convert.go` needs updating.

4. **Diffability.** `access_rights` is a map keyed by API ID, so the ordering is non-deterministic. The CLI schema uses an ordered list. Git diffs are clean.

## Duration system

`internal/policy/duration.go`

### Parsing rules

Accepted: `"30d"`, `"24h"`, `"1m"`, `"60s"`, `"60"`, `"0"`

Rejected: negative (`"-1d"`), fractional (`"1.5h"`), spaces (`"1 d"`), mixed units (`"1h30m"`), bare suffix (`"m"`), empty string.

### Why no mixed units?

Go's `time.ParseDuration` supports `"1h30m"`. We deliberately do not. Policy durations are configuration values, not arbitrary time spans. Mixed units create ambiguity in diffs (`"1h30m"` vs `"90m"`) and make it harder to enforce "largest clean unit" formatting on output. Every duration can be expressed in a single unit.

### Formatting: largest clean unit

`FormatDuration` converts seconds back to the largest unit that divides evenly:
- `86400` → `"1d"` (not `"24h"` or `"1440m"`)
- `3600` → `"1h"`
- `90` → `"90s"` (not `"1m30s"` — no mixed units)
- `0` → `"0"`

This ensures `get → edit → apply` round-trips produce minimal YAML diffs.

### Why a custom `Duration` type?

The `Duration` type in `pkg/types` is just `type Duration string` with a custom YAML unmarshaler. It exists because YAML treats bare `60` as an integer and `"60"` as a string. Without the custom unmarshaler, a policy file with `per: 60` (no quotes) would fail to parse into a string field. The unmarshaler reads the raw YAML node value regardless of tag type, so both `per: 60` and `per: "1m"` work.

## Selector resolution

`internal/policy/selector.go`

The apply pipeline needs to convert user-friendly API references into the hex IDs the Dashboard expects. This is the selector resolution step.

### Selector types

| Selector | Cardinality | Match semantics |
|----------|-------------|-----------------|
| `name` | Exactly 1 | Exact string match on API name |
| `listenPath` | Exactly 1 | Exact string match on listen path |
| `id` | Exactly 1 | Exact string match on API ID |
| `tags` | 1 or more | AND match — API must have ALL listed tags |

`name`, `listenPath`, and `id` are single-API selectors. They must resolve to exactly one API. `tags` is a multi-API selector — it expands to every API matching all the specified tags.

### Why restrict to exactly one selector per entry?

Each `AccessEntry` must have exactly one of `id`, `name`, `listenPath`, or `tags` set. This is validated in `internal/policy/validate.go` (the `selectorCount` function counts non-zero fields).

Allowing multiple selectors per entry (e.g. both `name` and `tags`) would create ambiguity: does the user want the intersection? The union? If they conflict, which wins? One selector per entry is unambiguous and easy to reason about.

### Fuzzy suggestions

When a name selector resolves to zero APIs, instead of just saying "not found", the CLI computes Levenshtein edit distance against all known API names and suggests the 3 closest matches:

```
no API found for name "user-api". Did you mean: users-api (a1b2c3), user-service (d4e5f6), user-mgmt (g7h8i9)
```

This is implemented with a single-row-optimized Levenshtein algorithm in `selector.go`. The fuzzy matching only kicks in for `name` selectors where typos are common — `listenPath` and `id` selectors fail without suggestions because those values are typically copied, not typed.

### The `ResolverAPI` decoupling

The selector resolver does not depend on `types.OASAPI` directly. Instead it takes `[]ResolverAPI`, a minimal struct with just `ID`, `Name`, `ListenPath`, and `Tags`. The CLI layer converts `[]*types.OASAPI` to `[]ResolverAPI` via the `toResolverAPIs` helper.

This decoupling means:
- The resolver is testable without HTTP mocks — just pass in a slice of `ResolverAPI`
- If the API type changes (e.g. Tyk adds new fields), the resolver doesn't need updating
- The resolver could work with any data source, not just the Dashboard API

## Validation: collect all errors

`internal/policy/validate.go`

Validation collects *all* errors before returning, rather than failing on the first one. This is deliberate:

```go
var errs types.ValidationErrors
// ... check each field, append to errs ...
return errs
```

The alternative — returning on the first error — forces users into an annoying cycle: fix one error, re-run, hit the next error, fix it, re-run. With multi-error collection, the user sees everything wrong at once and can fix it in a single pass.

### Validation layers

Validation happens in three layers during `apply`:

1. **Schema validation** (`validate.go`) — required fields, selector constraints. Runs offline, no network.
2. **Duration validation** (`validate.go`) — checks all duration fields parse correctly. Also offline.
3. **Selector resolution** (`selector.go`) — resolves selectors against the live API list. Requires network. Errors here include "not found" and "ambiguous" with suggestions.

The layers are ordered by cost: cheap checks first, network calls last. If schema validation fails, we never hit the Dashboard.

### The `ValidationError` type

Each error carries three fields:
- `Field` — JSON path like `"access[2]"` or `"rateLimit.per"`
- `Message` — human-readable description
- `Kind` — category: `"schema"`, `"duration"`, or `"selector"`

The `Kind` field exists so tooling or CI can filter errors by category.

## The apply pipeline

`internal/cli/policy.go` — `runPolicyApply`

The full pipeline for `tyk policy apply -f policy.yaml`:

```
Read file/stdin
    │
    ▼
Parse YAML → PolicyFile
    │
    ▼
Validate schema + durations + friendly ID format (offline)
    │  fail → exit 2 with all errors
    ▼
Fetch live API list from Dashboard
    │  fail → exit 1
    ▼
Resolve selectors against API list
    │  fail → exit 2 with suggestions
    ▼
Convert CLI schema → wire format (CLIToWire)
    │  Sets dp.ID = friendly id, leaves dp.MID empty
    │  fail → exit 2
    ▼
Resolve friendly ID: GET /api/portal/policies/{id}  (O(1))
    │
    ├── exists (200) → PUT /api/portal/policies/{id}
    └── not found (404) → POST /api/portal/policies
         │
         ▼
    Print confirmation to stderr
```

### Idempotent upsert

`apply` is a single command that creates or updates. There is no separate `create` or `update`. The existence check is a GET followed by either POST or PUT.

Why not a `--create-only` or `--update-only` flag? Because the primary use case is GitOps: you run `apply` in CI on every push, and it should converge to the desired state regardless of whether the policy already exists. Idempotent operations are safe to retry.

### Stdin support

`apply -f -` reads from stdin. This enables piping:

```bash
cat policy.yaml | envsubst | tyk policy apply -f -
```

The `-` convention follows `kubectl`, `docker`, and other CLI tools.

## Output conventions

All commands follow the same pattern established by `api.go`:

- **Human messages** go to **stderr** (summaries, prompts, "Policy created.")
- **Machine-readable data** goes to **stdout** (YAML, JSON, table rows)

This means you can pipe `tyk policy get gold > policy.yaml` and the summary line doesn't end up in the file.

The `--json` flag switches stdout output from the default format (YAML for `get`, table for `list`) to JSON. Stderr messages are unaffected.

## Exit codes

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | — | Success (or user-cancelled delete) |
| 1 | — | Network / server / unexpected error |
| 2 | `ExitBadArgs` | Validation error, resolution failure, bad input |
| 3 | `ExitNotFound` | Resource not found (get/delete with bad ID) |

Exit code 2 for validation errors (not 1) allows scripts to distinguish "your input is wrong" from "the server is down".

## Wire format quirks

### `allowed_urls` must be `[]`, not `null`

The Dashboard API rejects `"allowed_urls": null` but accepts `"allowed_urls": []`. Go's default JSON marshaling produces `null` for a nil slice. The custom `MarshalJSON` on `AccessRight` (`pkg/types/policy.go:142`) handles this:

```go
if ar.AllowedURLs == nil {
    a.AllowedURLs = []AllowedURL{}  // serialize as []
}
```

### `limit` must be explicit `null`

Similarly, per-API rate/quota limits (the `limit` field on `AccessRight`) must be explicitly `null` in the JSON, not omitted. The custom marshaler uses `json.RawMessage("null")` for nil limits.

### `_id` vs `id` — friendly ID resolution

The Dashboard stores two identifier fields on every policy:

- `_id` — the MongoDB document ID. A 24-character hex string generated by the Dashboard on creation. Handled invisibly by the Dashboard.
- `id` — the wire `id` field. Starting in Dashboard v5.12.0+, this is a first-class identifier that can be used directly for all CRUD operations.

The CLI passes the user's friendly `id` directly into the wire `id` field with no prefix or transformation:

```yaml
# User writes:
id: gold

# Wire format:
{"id": "gold", "name": "Gold Plan", ...}
```

**Key rules:**

1. The CLI **never sets or reads `_id`**. The Dashboard generates and manages it invisibly.
2. All API calls (GET, PUT, DELETE) use the `id` field value directly in the URL path.
3. On display (`get`, `list`), the CLI uses `id` directly. If `id` is empty (unmanaged policy), it falls back to `_id`.
4. User tags are untouched — the CLI does not inject managed metadata into tags.

### O(1) resolution

The Dashboard's `GET /api/portal/policies/{id}` endpoint (v5.12.0+) resolves the `{id}` parameter against the wire `id` field. This means the CLI can look up a policy by its friendly ID in a single API call:

```
GET /api/portal/policies/gold   -> 200 (found) or 404 (not found)
```

No list+scan, no pagination, no filtering. All single-policy operations (`get`, `delete`, `apply` existence check) are O(1).

### `access_rights` map key = API ID

The wire format stores access rights as `map[string]*AccessRight` where the key is the API ID. This is redundant — each `AccessRight` also has an `api_id` field. But the Dashboard requires the key. `CLIToWire` sets both.

## Reverse resolution (get)

When `tyk policy get` fetches a policy from the Dashboard, it converts wire format back to CLI schema. The interesting part is reverse-resolving API IDs to names.

The CLI fetches the current API list and builds a lookup table (`apiByID`). For each `access_rights` entry:
- If the API ID is found in the lookup, use `name` selector
- If the API ID is not found (e.g. API was deleted), fall back to `id` selector

This is **best-effort** and **non-fatal**. If the API list fetch fails entirely, all entries fall back to `id` selectors. The user still gets a valid, re-appliable policy file — just with IDs instead of names.

The access entries are sorted by API ID for deterministic YAML output across runs.

## Package structure

```
pkg/types/          Shared types (both formats, validation error type)
                    No business logic, just data structures + marshaling

internal/policy/    Pure domain logic (no HTTP, no CLI, no I/O)
  duration.go       Parse "30d" → 2592000, format 2592000 → "30d"
  validate.go       Schema + duration + friendly ID validation, error collection
  selector.go       API selector resolution + fuzzy suggestions
  convert.go        CLI schema ↔ wire format conversion
  resolve.go        API access entry resolution types (ResolverAPI, ResolvedAccess, ResolveRequest)

internal/client/    HTTP client methods (CRUD against Dashboard API)
  policy.go         ListPolicies, GetPolicy, CreatePolicy, UpdatePolicy, DeletePolicy

internal/cli/       Cobra command definitions (CLI layer, wiring)
  policy.go         Command handlers, pipeline orchestration, I/O
  root.go           Registers policy command in root tree
```

### Dependency direction

```
cli → client → types
cli → policy → types
```

`internal/policy` has zero dependency on `internal/client` or `internal/cli`. It depends only on `pkg/types`. This means all domain logic (parsing, validation, resolution, conversion) is testable with zero HTTP mocking — just pass in structs.

`internal/cli` is the integration layer that wires everything together: it calls the client, feeds results into the policy package, and handles I/O.

## Testing strategy

### Unit tests (no HTTP)

- `pkg/types/policy_test.go` — YAML/JSON round-trip, `Duration` unmarshaling, `MarshalJSON` nil handling
- `internal/policy/duration_test.go` — parse/format edge cases (negative, fractional, overflow, suffixes)
- `internal/policy/validate_test.go` — multi-error collection, field paths, selector count
- `internal/policy/selector_test.go` — resolve by name/path/id/tags, fuzzy suggestions, ambiguity, zero-match
- `internal/policy/convert_test.go` — CLI→wire→CLI round-trip, duration conversion, access rights mapping, ID pass-through

### Integration tests (with httptest)

- `internal/client/policy_test.go` — CRUD methods against `httptest.Server`, verifying request paths/methods/bodies
- `internal/cli/policy_test.go` — Full command tests: create Cobra commands, inject config via context, mock Dashboard with `httptest.Server`, capture stdout/stderr, verify exit codes

### Walking skeleton

`TestPolicyIntegration_FullLifecycle` in `policy_test.go` runs the complete lifecycle: list (empty) -> apply (create) -> list (shows policy with friendly ID) -> get (returns CLI schema with friendly ID) -> apply (update, idempotent) -> delete -> list (empty again). This is the acceptance test that proves all commands work together end-to-end, including O(1) resolution by `id`.

## Future considerations

### Per-API rate/quota limits

The wire format supports per-API `limit` overrides via the `AccessRight.Limit` field. The CLI schema currently does not expose this — `Limit` is always `nil` in the wire output. Adding per-API limits would mean extending `AccessEntry` with optional `rateLimit` and `quota` fields and updating `CLIToWire`/`WireToCLI`.

### URL-level access control

`AccessRight.AllowedURLs` supports restricting access to specific URL patterns within an API. Currently unused — always `[]`. If exposed, it would add an `allowedURLs` list to `AccessEntry`.

### Policy partials / inheritance

Some users want a "base policy" with overrides per tier (gold inherits from base, adds higher limits). This is not supported today. It could be implemented as a CLI-side feature (merge YAML files before apply) without changing the wire format.

### Multi-environment promotion

Policies reference APIs by name, which should be consistent across environments. But if API names differ between dev and staging, `apply` will fail with resolution errors. A future `--env-map` flag or mapping file could handle this.
