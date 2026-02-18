# Implementation Roadmap: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DESIGN -> DISTILL handoff
**Date**: 2026-02-18

---

## Simplest Solution Analysis

Before proposing a multi-step roadmap, consider simpler alternatives:

### Rejected Alternative 1: Single `apply` command only (no list/get/delete)

Users would apply policies from YAML but have no way to verify server state or discover existing policies from the CLI. This breaks the GitOps workflow (apply -> verify round-trip) documented in all user stories. Rejected: violates US-PM-01, US-PM-02, US-PM-04.

### Rejected Alternative 2: Pass-through to Dashboard API (no CLI schema, no selectors)

Users would write raw Dashboard JSON and the CLI would POST/PUT it directly. No duration parsing, no selector resolution, no human-friendly format. Rejected: eliminates the core value proposition (portable, human-friendly policy files) and every UAT scenario for US-PM-03.

---

## Phase 1: Core CRUD (Walking Skeleton)

**Stories**: US-PM-01, US-PM-02, US-PM-03, US-PM-04, US-PM-05
**New production files**: 7 (6 new + 1 modified)
**Steps**: 7
**Steps/files ratio**: 1.0

---

### Step 1: Policy Types

- **Description**: Define CLI schema types, wire types, and validation error types for policies.
- **Files**: `pkg/types/policy.go`
- **Acceptance Criteria**:
  - PolicyFile struct round-trips through YAML marshal/unmarshal
  - DashboardPolicy struct round-trips through JSON marshal/unmarshal
  - Duration fields accept both string and integer representations
  - Selector constraint enforced: exactly one of id/name/listenPath/tags per access entry

### Step 2: Policy Client CRUD

- **Description**: Add policy CRUD methods to existing Client struct using Dashboard policy endpoints.
- **Files**: `internal/client/policy.go`
- **Acceptance Criteria**:
  - ListPolicies returns paginated policy list from Dashboard
  - GetPolicy returns single policy by ID; 404 yields structured error
  - CreatePolicy sends POST with wire-format payload
  - UpdatePolicy sends PUT with wire-format payload
  - DeletePolicy sends DELETE; 404 yields structured error
- **Architectural Constraints**:
  - Methods on existing Client struct (reuse doRequest/handleResponse)
  - Endpoint paths as package constants

### Step 3: Duration Parser

- **Description**: Parse human-friendly duration strings to seconds and reverse-format seconds to best human unit.
- **Files**: `internal/policy/duration.go`
- **Acceptance Criteria**:
  - Parses s/m/h/d suffixes and plain integers to seconds
  - Rejects fractional, negative, and invalid inputs
  - Formats seconds back to largest clean unit
  - Zero returns zero (special case: no expiry)

### Step 4: Schema Validation and Selector Resolution

- **Description**: Validate policy YAML schema and resolve API selectors to Dashboard IDs.
- **Files**: `internal/policy/validate.go`, `internal/policy/selector.go`
- **Acceptance Criteria**:
  - Validates required fields, types, duration formats, selector format
  - Collects all validation errors with field paths before returning
  - Name/listenPath/id selectors resolve to exactly one API or fail
  - Tags selector resolves to one or more APIs or fails
  - Zero-match failures include fuzzy suggestions (top 3 by edit distance)
  - Ambiguous-match failures include candidate list with IDs

### Step 5: Wire Format Conversion

- **Description**: Convert between CLI schema and Dashboard wire format in both directions.
- **Files**: `internal/policy/convert.go`
- **Acceptance Criteria**:
  - CLI-to-wire maps all fields per data model spec
  - Wire-to-CLI reverse-maps with best-effort API name resolution
  - Round-trip: apply then get produces equivalent policy content
  - Unresolvable API IDs fall back to id selector in output

### Step 6: CLI Commands (list + get + apply + delete + init)

- **Description**: Cobra command tree for all Phase 1 policy subcommands, wiring types, client, and policy logic.
- **Files**: `internal/cli/policy.go`
- **Acceptance Criteria**:
  - `list` displays paginated table (ID, Name, APIs, Tags) to stdout; header to stderr
  - `get` shows summary to stderr, CLI schema YAML to stdout; `--json` for JSON output
  - `apply -f` validates, resolves selectors, converts, and upserts idempotently
  - `apply -f -` reads from stdin
  - `delete` confirms interactively unless `--yes`; `--json` for structured output
  - `init` prompts for ID/name and writes scaffold YAML; warns on existing file
  - Not-found errors return exit code 3
  - Validation/selector errors return exit code 2
- **Architectural Constraints**:
  - Follow output convention: stderr for humans, stdout for data
  - Reuse GetConfigFromContext, GetOutputFormatFromContext, ExitError

### Step 7: Root Command Registration

- **Description**: Register policy command group in the root command.
- **Files**: `internal/cli/root.go` (1-line modification)
- **Acceptance Criteria**:
  - `tyk policy --help` shows all subcommands
  - `tyk --help` lists `policy` alongside `api` and `config`

---

## Phase 2: Cross-referencing and Helpers (Future)

**Stories**: US-PM-06, US-PM-07
**Depends on**: Phase 1 complete
**No architectural changes needed** -- builds on existing types, client, and selector infrastructure.

### Step 8: Who-Uses Command

- **Description**: Show which policies reference a given API.
- **Files**: `internal/cli/api.go` (add who-uses subcommand)
- **Acceptance Criteria**:
  - Accepts API ref by name, ID, or listenPath
  - Lists referencing policies in table format
  - Exit code 3 when API ref not found
  - `--json` for machine output

### Step 9: Bind and Unbind Commands

- **Description**: Quick-add or quick-remove an API from a policy's access rights.
- **Files**: `internal/cli/policy.go` (add bind/unbind subcommands)
- **Acceptance Criteria**:
  - Bind adds API to policy via GET + modify + PUT
  - Unbind removes API from policy via GET + modify + PUT
  - Duplicate bind returns informative error (exit 2)
  - Missing unbind returns informative error (exit 2)

---

## Implementation Notes for Crafter

1. **Walking skeleton**: Steps 1-2-6 (types + client + CLI for `list`) can produce a vertical slice quickly. Extend with steps 3-5 to enable `apply`.
2. **Test-first**: Each step has testable units. The `internal/policy/` package is pure logic (no cobra, no HTTP) -- highly unit-testable.
3. **Table helpers**: `truncateWithEllipsis` and `computeTableLayout` in `api.go` are unexported. The crafter decides whether to export them, duplicate them, or extract to a shared file.
4. **Endpoint confirmation**: The Dashboard policy endpoint prefix (`/api/portal/policies` vs `/api/policies`) should be confirmed early in Step 2. The architecture isolates this to a single constant.
5. **ListAPIsDashboard pagination**: For selector resolution, the crafter needs a "list all APIs" helper that fetches all pages. This could be a loop calling `ListAPIsDashboard` with incrementing page numbers until an empty page is returned.
