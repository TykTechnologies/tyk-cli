# Requirements: OAS Versioning Support

## Overview

Native OAS version management in the Tyk CLI, enabling developers to list, inspect, create, and manage API versions through scriptable, composable commands that integrate with existing apply workflows.

## Functional Requirements

### FR-1: List API Versions
The CLI must list all versions of a given OAS API, showing version names and which is the default. Must support `--json` output for scripting.

### FR-2: Get Specific Version
The existing `tyk api get` command must support `--version-name` flag to retrieve a specific version's OAS document. Already partially implemented -- needs error handling for missing versions with available-versions hint.

### FR-3: Create New Version
The CLI must create a new version for an existing API from a local OAS file. Must accept `--api-id`, `-f` (OAS file), `--version-name`, and optional `--set-default`. Must reject duplicate version names with exit code 4.

### FR-4: Switch Default Version
The CLI must switch which version serves as the default for an API. Must show previous and new default. Must be a no-op (exit 0) when the target is already the default.

### FR-5: Apply Integration
The existing `tyk api apply` and `tyk apply -f` commands must support version operations:
- Single file apply with `--version-name` and `--base-api-id` flags
- Batch apply recognizes `x-tyk-api-gateway.info.versioning` extension to auto-detect version files
- Dry-run reports version operations accurately

## Non-Functional Requirements

### NFR-1: CLI Conventions
All commands follow established patterns: stdout for machine data, stderr for human messages, exit codes 0/1/2/3/4, `--json` flag support.

### NFR-2: Composability
Output of one command can feed into another. `list --json | jq` for scripting. Version names from list can be used in get/switch-default.

### NFR-3: Error Recoverability
Error messages include actionable context: version-not-found errors list available versions. Conflict errors suggest the correct command (apply instead of create).

## Constraints

- Must use existing Dashboard API endpoints (no new backend endpoints)
- Must use existing client methods: `ListOASAPIVersions`, `SwitchDefaultVersion`, `GetOASAPI`
- Must use existing `CreateOASAPIRequest` type (has `BaseAPIID`, `NewVersionName`, `SetDefault` fields)
- Must follow cobra command structure established in `internal/cli/api.go`
- Placeholder commands already exist at `tyk api versions {list,create,switch-default}`

## Dependencies

- `internal/client/client.go`: Client methods for version operations (exist)
- `pkg/types/api.go`: Types for version data (exist)
- `internal/oas/transform.go`: OAS parsing utilities (exist)
- `internal/cli/apply.go`: Batch apply framework (exists, needs extension for version awareness)
