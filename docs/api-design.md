# API Management Design for Tyk CLI

This document describes the design of OAS-first API management in the Tyk CLI. It covers the architecture, workflows, and design decisions behind the `tyk api` command group.

## Goals

- Treat OpenAPI Specification (OAS) files as the source of truth for API definitions.
- Support multiple ingestion paths: quick creation from flags, import from external OAS specs, and declarative apply from Tyk-enhanced OAS files.
- Provide an idempotent `apply` flow for CI/CD and GitOps workflows.
- Separate human-readable output (stderr) from machine-parseable data (stdout) so commands compose cleanly with pipes.
- Keep the client layer thin: the Dashboard API is the authority; the CLI is a UX layer on top.

## Resource Model

- Kind: OAS API (OpenAPI 3.0 document with optional `x-tyk-api-gateway` extensions).
- Identity:
  - `x-tyk-api-gateway.info.id` (string, assigned by the Dashboard on creation or declared in file for idempotent upsert).
  - `info.title` (human-readable name, extracted from OAS `info` section).
- Key fields extracted from extensions:
  - `x-tyk-api-gateway.info.name` — API name as known to the Gateway.
  - `x-tyk-api-gateway.server.listenPath.value` — the path the Gateway listens on.
  - `x-tyk-api-gateway.upstream.url` — the upstream service URL.
  - `x-tyk-api-gateway.info.state.active` — whether the API is active.
- Versioning: APIs support multiple versions; each version can carry its own OAS document. The `default_version` field controls which version responds when no version header is sent.

## Three Ingestion Paths

The CLI provides three distinct ways to get an API into the Dashboard, each suited to a different stage of the workflow:

### 1. `tyk api create` — Quick scaffold

Creates an API from CLI flags. Internally generates a minimal OAS document with Tyk extensions and posts it to the Dashboard.

```bash
tyk api create --name "User Service" --upstream-url https://users.api.com
```

When `--listen-path` is omitted, the CLI auto-generates one from the name (e.g. `"User Service"` becomes `/user-service/`). This uses a slugification algorithm in `internal/oas/transform.go` that lowercases, replaces non-alphanumeric characters with hyphens, and wraps with `/`.

Best for: bootstrapping a new API quickly during local development.

### 2. `tyk api import-oas` — External spec ingestion

Imports a clean OpenAPI specification (no Tyk extensions) from a file or URL. The CLI auto-generates `x-tyk-api-gateway` extensions (name, listen path, upstream, active state) and creates a new API. The API always gets a new Dashboard-assigned ID.

```bash
tyk api import-oas --file petstore.yaml
tyk api import-oas --url https://api.example.com/openapi.json
```

The import path strips any pre-existing `x-tyk-api-gateway.info.id` to guarantee a fresh API. This is intentional: `import-oas` is a "create new" operation, not an upsert.

Best for: onboarding third-party or externally-authored API specs.

### 3. `tyk api apply` — Declarative upsert (GitOps)

Applies a Tyk-enhanced OAS file idempotently. This is the primary GitOps command.

```bash
tyk api apply --file enhanced-api.yaml    # From file
tyk api apply --file -                     # From stdin (piping)
```

Behavior depends on whether `x-tyk-api-gateway.info.id` is present:

- **ID present**: Check if the API exists on the Dashboard.
  - Exists: `PUT` (update).
  - Not found: `POST` (create with the declared ID) — enables idempotent upsert.
- **ID absent**: Always creates a new API (Dashboard assigns the ID).

The file **must** contain `x-tyk-api-gateway` extensions. If it doesn't, `apply` rejects the file with exit code 2 and suggests using `import-oas` or `update-oas` instead. This is a deliberate guardrail: `apply` is for fully-specified configurations, not for auto-enhancing plain specs.

Best for: CI/CD pipelines, GitOps, and infrastructure-as-code workflows.

### Why three paths?

A single `apply` command could theoretically handle all cases (detect missing extensions, generate them, upsert). We split into three for clarity of intent:

1. **`create`** — imperative, quick, local-dev-friendly. No file needed.
2. **`import-oas`** — one-time ingestion of external specs. Always creates new.
3. **`apply`** — declarative, idempotent, CI-safe. Requires explicit Tyk configuration.

Each command has clear, non-overlapping semantics. Users don't need to reason about "what will happen if my file does/doesn't have extensions?" — the command name tells them.

### `tyk api update-oas` — Surgical spec update

Updates only the OAS portion of an existing API while preserving all Tyk-specific middleware and configuration. Takes a clean OpenAPI spec and merges it with the existing `x-tyk-api-gateway` extensions on the server.

```bash
tyk api update-oas <api-id> --file new-spec.yaml
tyk api update-oas <api-id> --url https://api.example.com/openapi.json
```

Best for: updating the API contract (paths, schemas, descriptions) without touching gateway configuration.

## OAS Extension Handling

### Auto-generation (`internal/oas/transform.go`)

When a plain OAS document needs Tyk extensions (for `create` and `import-oas`), the CLI generates minimal extensions:

- `info.name` from OAS `info.title`.
- `info.state.active = true`.
- `upstream.url` from the first entry in OAS `servers`.
- `server.listenPath.value` from slugified title; `strip = true`.

No ID is set — the Dashboard assigns one on creation.

### Extension detection

`HasTykExtensions` simply checks for the presence of the `x-tyk-api-gateway` key. It does not validate the extension structure. This is intentional: the Dashboard performs authoritative validation. The CLI only checks enough to route to the correct code path.

### ID extraction

`ExtractAPIIDFromTykExtensions` navigates the nested map `x-tyk-api-gateway.info.id`. It returns `("", false)` for any structural issue (missing keys, wrong types, empty string). The caller decides whether a missing ID means "create new" or "error".

## CLI Commands

Top-level: `tyk api`.

| Command | Args / Flags | Behavior |
|---------|-------------|----------|
| `list` | `--page N`, `-i` (interactive) | Paginated listing; interactive mode with arrow-key navigation |
| `get <id>` | `--version-name`, `--oas-only` | Fetch API; default output is YAML OAS to stdout |
| `create` | `--name`, `--upstream-url`, `--listen-path`, `--version-name`, `--custom-domain`, `--description` | Quick-create from flags |
| `import-oas` | `--file` or `--url` | Import external OAS spec |
| `apply` | `--file` (required), `--version-name`, `--set-default` | Declarative upsert |
| `update-oas <id>` | `--file` or `--url` | Update OAS portion only |
| `delete <id>` | `--yes` | Delete with confirmation (or skip with `--yes`) |

Planned (post-v0):
- `tyk api versions list|create|switch-default` — version management.
- `tyk api convert` — format conversion between OAS and Tyk API definition.

## Output Conventions

All commands follow the same pattern:

- **Human messages** go to **stderr** (summaries, prompts, banners, navigation hints).
- **Machine-readable data** goes to **stdout** (YAML, JSON, table rows).

This means `tyk api get <id> --oas-only > api.yaml` produces a clean file without summaries leaking in.

The `--json` output format (set via `--output json` on the root command) switches stdout output to JSON. Stderr messages are unaffected.

### Interactive list mode

`tyk api list -i` enters a full-screen interactive mode with:

- Terminal raw mode for keystroke capture.
- Arrow keys / A/D for page navigation, R for refresh, Q to quit.
- Adaptive table layout that adjusts column widths to terminal size.
- Cursor hiding during repainting to prevent flicker.

Interactive mode is incompatible with JSON output (fails fast with an error).

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (or user-cancelled delete) |
| 1 | Network / server / unexpected error |
| 2 | Validation error, bad input, missing extensions |
| 3 | Resource not found (get/delete with bad ID) |
| 4 | Conflict (e.g., duplicate API on create) |

## File Loading

### Local files

Files are loaded via `internal/filehandler/filehandler.go`. The handler:

- Resolves relative paths to absolute.
- Detects format (JSON or YAML) by content, not extension.
- Returns a parsed `map[string]interface{}` representation.

### Remote URLs

`loadOASFromURL` fetches the document over HTTP(S) with a 30-second timeout. It attempts JSON parsing first, then YAML. This order is intentional: JSON parsing is strict (rejects YAML-only constructs), while YAML parsing accepts JSON as a subset.

### Stdin

`apply -f -` reads from stdin, enabling piping:

```bash
cat api.yaml | envsubst | tyk api apply -f -
```

The `-` convention follows `kubectl`, `docker`, and other CLI tools.

## Client Layer (`internal/client/client.go`)

### Dashboard API endpoints

| Operation | Method | Path |
|-----------|--------|------|
| List (OAS) | GET | `/api/apis/oas?p={page}` |
| List (aggregate) | GET | `/api/apis?p={page}` |
| Get | GET | `/api/apis/oas/{apiId}` |
| Create | POST | `/api/apis/oas` |
| Update | PUT | `/api/apis/oas/{apiId}` |
| Delete | DELETE | `/api/apis/oas/{apiId}` |
| List versions | GET | `/api/apis/oas/{apiId}/versions` |

### Two list endpoints

The CLI uses `ListAPIsDashboard` (the `/api/apis` aggregate endpoint) rather than `ListOASAPIs` (the `/api/apis/oas` endpoint) for the `list` command. The aggregate endpoint returns both classic and OAS APIs in a unified format, providing broader compatibility across Dashboard versions. The response structure differs:

- `/api/apis` returns `{ apis: [{ api_definition: { api_id, name, proxy: { listen_path } } }] }`.
- `/api/apis/oas` returns `{ apis: [<OASAPI>] }`.

The client maps the aggregate response into the same `[]*types.OASAPI` slice used by the CLI layer.

### Create/Update pattern

Both `CreateOASAPI` and `UpdateOASAPI` follow a two-step pattern:

1. Send the OAS document (POST or PUT).
2. The Dashboard returns only basic info (`{ Status, ID }`).
3. Immediately `GET` the full API to return complete metadata to the caller.

This ensures the caller always gets a fully-populated `*types.OASAPI`, regardless of what the Dashboard returns in the mutation response.

### Error handling

HTTP responses with status >= 400 are parsed into `*types.ErrorResponse`. The client attempts JSON parsing first; if the body isn't JSON, the raw status text and body become the error message.

For the `apply` upsert flow, the CLI distinguishes between "not found" (create path) and actual errors. It checks both typed `*types.ErrorResponse` with status 404 and string-based fallbacks for Dashboards that return 400 with "could not retrieve api" messages.

## Type System (`pkg/types/api.go`)

The type system is deliberately minimal:

- `OASAPI` — the CLI's view of an API. Contains extracted metadata (ID, name, listen path, upstream, versions) plus the raw OAS document as `map[string]interface{}`.
- `APIVersion` — per-version data including its own OAS document.
- `CreateOASAPIRequest` / `UpdateOASAPIRequest` — wire-format request types.
- `ErrorResponse` — structured error with status code, message, and optional details.

The `OAS` field is `map[string]interface{}` rather than a typed struct. This is deliberate: OAS documents are extensible and vary widely. Parsing into a strict struct would lose extensions, custom fields, and vendor properties. The CLI only needs to read/write a few known paths (`x-tyk-api-gateway`, `info`, `servers`); everything else passes through untouched.

## Upsert Semantics in `apply`

The apply pipeline for `tyk api apply -f api.yaml`:

```
Read file/stdin
    |
    v
Parse YAML/JSON -> map[string]interface{}
    |
    v
Check for x-tyk-api-gateway extensions
    |  missing -> exit 2 with guidance
    v
Extract API ID from extensions
    |
    +-- ID present -> Check if API exists (GET)
    |       |
    |       +-- exists -> PUT (update)
    |       +-- not found -> POST (create with declared ID)
    |
    +-- ID absent -> POST (create, Dashboard assigns ID)
         |
         v
    Print result (created/updated) to stderr
    Print API details to stdout
```

### Why not a separate `create` and `update`?

Because the primary use case is GitOps: you commit an API file, CI runs `apply` on every push, and it converges to the desired state. Having separate commands forces the user to track whether the API already exists — state that belongs to the server, not to the user's workflow.

### Conflict handling

Exit code 4 is reserved for conflicts (HTTP 409). This can happen when the Dashboard detects a listen-path collision or other uniqueness violation. The error message from the Dashboard is surfaced directly.

## Package Structure

```
pkg/types/           Shared types (OASAPI, request/response types, ErrorResponse)
                     No business logic, just data structures

internal/oas/        OAS document utilities
  transform.go       Extension detection, ID extraction, auto-generation, slug generation

internal/filehandler/ File loading
  filehandler.go     Loads JSON/YAML files into map[string]interface{}

internal/client/     HTTP client (CRUD against Dashboard API)
  client.go          All OAS API operations, response parsing, error mapping

internal/cli/        Cobra command definitions (CLI layer, wiring)
  api.go             Command handlers, output formatting, interactive list
```

### Dependency direction

```
cli -> client -> types
cli -> oas    -> (stdlib only)
cli -> filehandler -> (stdlib only)
```

`internal/oas` and `internal/filehandler` have no dependency on `internal/client` or `internal/cli`. They are pure utility packages.

## Testing Strategy

### Unit tests

- `internal/oas/transform_test.go` — extension detection, ID extraction, slug generation, auto-enhancement edge cases.
- `internal/filehandler/filehandler_test.go` — file loading for JSON, YAML, invalid content.

### Integration tests (with httptest)

- `internal/cli/api_list_test.go` — list command with pagination, JSON output, empty results.
- `internal/cli/api_get_test.go` — get command with version selection, OAS-only mode, not-found handling.
- `internal/cli/api_create_test.go` — create command with various flag combinations, conflict handling.
- `internal/cli/api_import_update_test.go` — import-oas and update-oas workflows.
- `internal/cli/api_interactive_test.go` — interactive list mode keystroke handling.
- `internal/client/client_test.go` — HTTP client methods against `httptest.Server`.

### Testing pattern

CLI tests create Cobra commands, inject config via context, mock the Dashboard with `httptest.Server`, and capture stdout/stderr separately. Exit codes are verified via the `ExitError` type.

## Relationship to Other Features

### Policies

Policies reference APIs by ID in their `access_rights` map. The policy system's selector resolution (`internal/policy/selector.go`) uses `ListAPIsDashboard` to map human-friendly selectors (name, listen path, tags) to API IDs. This makes policies portable: policy files reference APIs by name, and the CLI resolves to IDs at apply time.

`tyk api who-uses <api-ref>` (planned) will list policies that reference a given API, surfacing dependencies before changes or deletions.

### Credentials

Credentials (API keys, OAuth clients) bind to policies, which in turn reference APIs. The credential system does not interact with the API management layer directly — the policy is the intermediary.

### Portal

Portal products reference APIs by `api_id`. When publishing an API to the portal, the product spec includes the Dashboard API ID. The portal design proposes resolving API selectors (name, listen path) to IDs via the Dashboard, similar to how policy selectors work.

## Future Considerations

### API versioning commands

The version subcommands (`versions list|create|switch-default`) are scaffolded but not yet implemented. They will use the existing `/api/apis/oas/{apiId}/versions` endpoint. Design considerations:

- `versions create` should accept either a new OAS file or clone the current default version.
- `switch-default` is a lightweight PATCH operation.
- Version deletion is intentionally not planned (versions are immutable records).

### `tyk api convert`

Convert between OAS and Tyk classic API definition formats. Useful for migration from classic APIs to OAS-first.

### Diff and dry-run

A `tyk api diff --file api.yaml` command that shows what would change without applying. Similar to `kubectl diff`. Requires fetching the current state and computing a structural diff against the file.

### Bulk apply

`tyk api apply --dir apis/` to apply all OAS files in a directory. Ordering would be alphabetical; dependencies between APIs (if any) would need explicit sequencing or a manifest file.

### API validation / linting

Client-side validation of OAS documents before sending to the Dashboard. Could use standard OAS validators or Tyk-specific rules (e.g., listen path format, upstream URL reachability).

## Open Questions

- Confirm exact version management endpoints and their behavior across Dashboard versions.
- Decide on the `convert` command's scope: full round-trip fidelity or best-effort migration?
- Consider whether `apply` should support `--dry-run` natively (Dashboard may not support a dry-run mode).
- Evaluate whether to add `--prune` support to remove APIs not present in a directory of files.

---

With this design, OAS APIs are first-class, version-controllable resources. The three ingestion paths cover the full spectrum from quick prototyping to full GitOps, while the idempotent `apply` command ensures CI/CD pipelines converge safely to the desired state.
