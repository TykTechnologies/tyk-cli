# Evolution Record: policies-mgmt

**Date:** 2026-02-18
**Project ID:** policies-mgmt
**Status:** COMPLETE

## Feature Summary

Policy management CLI commands for the Tyk Dashboard. Enables declarative management of security policies through `tyk policy` with five subcommands: `list`, `get`, `apply`, `delete`, and `init`. Policies are authored as CLI-schema YAML files with human-friendly duration formats and API selectors (by name, listen path, ID, or tags), then converted to Dashboard wire format on apply.

## Phase Breakdown

### Phase 01 -- Foundation: Types, Client, and Policy Logic (Steps 01-01 to 01-03)

| Step | Description | Commit |
|------|-------------|--------|
| 01-01 | CLI schema types (PolicyFile, PolicyMetadata, PolicySpec, AccessEntry) and Dashboard wire types (DashboardPolicy, AccessRight). Duration fields accept both string and integer YAML input. | `eb15192` |
| 01-02 | CRUD client methods (ListPolicies, GetPolicy, CreatePolicy, UpdatePolicy, DeletePolicy) using existing doRequest/handleResponse pattern against `/api/portal/policies`. | `8e4cfd6` |
| 01-03 | Duration parser (s/m/h/d suffixes), schema validator (multi-error collection with field paths), selector resolver (name/listenPath/id/tags with fuzzy suggestions), bidirectional CLI-to-wire converter. | `1711062` |

### Phase 02 -- CLI Commands: list, get, apply, delete (Steps 02-01 to 02-04)

| Step | Description | Commit |
|------|-------------|--------|
| 02-01 | `tyk policy list` with paginated table output (ID, Name, APIs, Tags). JSON output mode. Registration in root command tree. | `f1b38a1` |
| 02-02 | `tyk policy get <id>` with wire-to-CLI conversion, reverse-resolving API IDs to name selectors. YAML/JSON output. | `ee98fcc` |
| 02-03 | `tyk policy apply -f <file>` with full pipeline: YAML load, schema validation, selector resolution against live API list, duration parsing, wire conversion, idempotent upsert. Stdin support with `-f -`. | `10c0154` |
| 02-04 | `tyk policy delete <id>` with existence verification, confirmation prompt (`--yes` to skip), and JSON output mode. | `355fb71` |

### Phase 03 -- Integration: init scaffold and end-to-end validation (Step 03-01)

| Step | Description | Commit |
|------|-------------|--------|
| 03-01 | `tyk policy init` scaffold generator (prompts for ID and name, writes `policies/{id}.yaml`). Full walking skeleton integration test: list empty, apply, list, get, delete. | `5ad59a1` |

## Key Decisions

- **CLI-schema vs wire-format separation:** Policies are authored in a human-friendly CLI schema (string durations, API selectors by name/path/tags) and converted to Dashboard wire format (integer seconds, API IDs in access_rights map) at apply time. This keeps YAML files readable and version-controllable.
- **Multi-error validation:** Schema validation collects all errors with field paths before returning, so users fix all issues in one pass rather than iterating on one error at a time.
- **Fuzzy suggestions on selector miss:** When a selector resolves to zero APIs, the top 3 fuzzy matches are returned as suggestions, reducing user friction.
- **Idempotent upsert for apply:** `apply` creates if the policy ID does not exist on the server, updates if it does. No separate create/update commands needed.
- **Offline init:** The `init` command generates valid scaffold YAML with no Dashboard connectivity required.

## Quality Gates

| Gate | Result | Detail |
|------|--------|--------|
| 5-phase TDD cycle (all 8 steps) | PASS | Every step completed PREPARE, RED_ACCEPTANCE, RED_UNIT, GREEN, COMMIT phases |
| L1-L4 refactoring | PASS | Completed post-implementation |
| Adversarial review | APPROVED | -- |
| Mutation testing | PASS | 81.82% efficacy (threshold: 80%) |
| Integrity verification | PASS | All production and test files verified |

## Files Created/Modified

### Production Files

| File | Description |
|------|-------------|
| `pkg/types/policy.go` | CLI schema types and Dashboard wire types |
| `internal/client/policy.go` | CRUD client methods for Dashboard policy API |
| `internal/policy/duration.go` | Duration parsing and formatting (s/m/h/d) |
| `internal/policy/validate.go` | Schema validation with field-path error collection |
| `internal/policy/selector.go` | API selector resolution with fuzzy suggestions |
| `internal/policy/convert.go` | Bidirectional CLI-to-wire format conversion |
| `internal/cli/policy.go` | Cobra command tree (list, get, apply, delete, init) |
| `internal/cli/root.go` | Modified to register policy command |

### Test Files

| File | Description |
|------|-------------|
| `pkg/types/policy_test.go` | Type round-trip and duration unmarshalling tests |
| `internal/client/policy_test.go` | Client CRUD tests with httptest server |
| `internal/policy/duration_test.go` | Duration parse/format edge cases |
| `internal/policy/validate_test.go` | Validation error collection and field paths |
| `internal/policy/selector_test.go` | Selector resolution and fuzzy suggestion tests |
| `internal/policy/convert_test.go` | CLI-to-wire round-trip conversion tests |
| `internal/cli/policy_test.go` | CLI command tests including integration walking skeleton |

## Metrics

| Metric | Value |
|--------|-------|
| Total test cases | 116 |
| Test packages | 4 (pkg/types, internal/client, internal/policy, internal/cli) |
| Mutation efficacy | 81.82% |
| Implementation steps | 8 |
| Commits | 8 (one per step) |
| Execution time | ~37 minutes (12:43 to 13:20 UTC) |
| TDD phases executed | 40 (8 steps x 5 phases, all PASS) |
