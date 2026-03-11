---
title: API Versioning Design
---

# API Versioning Design

Design decisions and architecture for OAS version management in the Tyk CLI.

## Overview

Each Tyk OAS API can have multiple versions. Each version has its own OAS document and can optionally override the upstream URL, listen path, and custom domain. One version is always marked as the default.

The CLI communicates with the Tyk Dashboard REST API for all version operations. No local state is stored.

## Data flow

### List versions

```
tyk api versions list --api-id abc123
  -> client.ListOASAPIVersions(ctx, "abc123")
  -> GET /api/apis/oas/abc123/versions
  <- {versions: ["v1","v2","v3"], default: "v1"}
  <- stderr (human table) or stdout (JSON)
```

### Create version

```
tyk api versions create --api-id abc123 -f v3.yaml --version-name v3
  -> filehandler.LoadFile("v3.yaml")
  -> client.ListOASAPIVersions(ctx, "abc123")     [conflict check]
  -> client.CreateOASAPIVersion(ctx, oasJSON, "abc123", "v3", false)
  -> POST /api/apis/oas?base_api_id=abc123&new_version_name=v3
  <- stderr (summary + hint) or stdout (JSON)
```

### Switch default

```
tyk api versions switch-default --api-id abc123 --version-name v3
  -> client.ListOASAPIVersions(ctx, "abc123")     [validate + get current]
  -> client.SwitchDefaultVersion(ctx, "abc123", "v3")
  -> PATCH /api/apis/oas/abc123  {set_default_version: "v3"}
  <- stderr (previous/new) or stdout (JSON)
```

### Batch apply with version files

```
tyk apply -f apis/
  -> scanDirectory("apis/")
  -> classifyContent() detects x-tyk-api-gateway.info.versioning
  -> sortFiles(): [policies, base APIs, version files]
  -> POST /api/apis/oas?base_api_id=X&new_version_name=Y
```

## Versioning extension schema

Version files declare their relationship to a base API using an embedded extension:

```yaml
x-tyk-api-gateway:
  info:
    name: "Petstore"
    versioning:
      base_api_id: "abc123"
      version_name: "v3"
      set_default: false
```

This is only consumed by `tyk apply -f` for classification. Direct commands use `--api-id` and `--version-name` flags.

## Files changed

| File | Change |
|------|--------|
| `internal/cli/version_errors.go` | New — error helpers (not-found, conflict, API-not-found) |
| `internal/oas/transform.go` | Added `VersioningMetadata` type and `ExtractVersioningMetadata()` |
| `internal/client/client.go` | Added `CreateOASAPIVersion()` method |
| `internal/cli/api.go` | Replaced 3 placeholders with real commands, wired versions subcommand |
| `internal/cli/apply.go` | Extended classification, ordering, dry-run, and apply for version files |

## Architecture decisions

### ADR-001: Pre-check pattern for version validation

**Decision**: Call `ListOASAPIVersions` before mutations (create, switch-default) to validate state.

**Why**: Dashboard error responses for version conflicts are inconsistent. Pre-checking is reliable and enables rich error messages listing available versions.

**Trade-off**: One extra HTTP round-trip per command (~50ms), acceptable for a CLI.

### ADR-002: Embedded versioning metadata for batch apply

**Decision**: Version files declare their relationship in `x-tyk-api-gateway.info.versioning` rather than relying on file naming or an external manifest.

**Why**: Self-contained — each file carries its own version parameters. No fragile naming conventions or extra files to maintain.

**Trade-off**: Adds a Tyk-specific extension sub-section, but files already have `x-tyk-api-gateway` so portability is already limited.

### ADR-003: Separate error helpers file

**Decision**: Version-specific error helpers live in `internal/cli/version_errors.go` rather than in `api.go`.

**Why**: `api.go` is ~1500 lines. Version errors are a distinct cross-cutting concern shared across all version commands.
