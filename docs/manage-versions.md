---
title: Manage API Versions
nav_order: 5
---

# Manage API Versions

Tyk OAS APIs support multiple versions, each with its own OpenAPI spec, upstream URL, and listen path. The CLI lets you list, create, and switch versions from the terminal or CI/CD pipelines.

## Quick start

```bash
# See what versions exist
tyk api versions list --api-id abc123

# Create a new version from an OAS file
tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3

# Switch default traffic to the new version
tyk api versions switch-default --api-id abc123 --version-name v3
```

## Commands

### `tyk api versions list`

List all versions for an API, with the default version marked.

```bash
tyk api versions list --api-id abc123          # Human table to stderr
tyk api versions list --api-id abc123 --json   # JSON to stdout
```

Example output:

```
* v1
  v2
  v3

3 version(s) found. Default: v1
```

JSON output:

```json
{
  "api_id": "abc123",
  "default": "v1",
  "versions": ["v1", "v2", "v3"]
}
```

### `tyk api versions create`

Create a new version linked to an existing base API from an OAS file.

```bash
tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3
tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2 --set-default
tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3 --json
```

| Flag | Required | Description |
|------|----------|-------------|
| `--api-id` | Yes | Base API to add the version to |
| `-f` | Yes | Path to OAS file |
| `--version-name` | Yes | Name for the new version (e.g. `v3`) |
| `--set-default` | No | Set as default version immediately |

The command checks for duplicate version names before creating. If the version already exists, it returns exit code 4 with a suggestion to use `tyk api apply` instead.

When `--set-default` is not used, the output includes a hint showing the `switch-default` command.

### `tyk api versions switch-default`

Switch which version serves as the default.

```bash
tyk api versions switch-default --api-id abc123 --version-name v3
tyk api versions switch-default --api-id abc123 --version-name v3 --json
```

The command validates the target version exists before switching. If the target is already the default, it exits 0 with an informational message and makes no API call.

JSON output:

```json
{
  "action": "switched",
  "api_id": "abc123",
  "previous_default": "v1",
  "new_default": "v3"
}
```

## Batch apply with versions

Version files work with `tyk apply -f` for GitOps workflows. Add a `versioning` section to the Tyk extension to declare the version relationship:

```yaml
openapi: "3.0.3"
info:
  title: Petstore
  version: "3.0.0"
x-tyk-api-gateway:
  info:
    name: Petstore
    versioning:
      base_api_id: "abc123"
      version_name: "v3"
      set_default: false
servers:
  - url: https://api.petstore.io/v3
paths: {}
```

| Field | Required | Description |
|-------|----------|-------------|
| `base_api_id` | Yes | ID of the parent API this version belongs to |
| `version_name` | Yes | Version identifier, must be unique within the parent |
| `set_default` | No | Set as default on creation (default: `false`) |

### Ordering

When applying a directory, files are processed in this order:

1. Policies
2. Base APIs (OAS files with `x-tyk-api-gateway` but no `versioning` section)
3. Version files (OAS files with `versioning` section)

This ensures the base API exists before its versions are created.

### Example directory layout

```
apis/
  petstore-base.yaml       # Base API (applied first)
  petstore-v3.yaml         # Version file with versioning metadata (applied second)
  policies/
    gold.yaml              # Policy (applied first)
```

```bash
tyk apply -f apis/                    # Apply all
tyk apply -f apis/ --dry-run          # Preview: "would create version: v3 of abc123"
tyk apply -f apis/ --json             # JSON results include version_name field
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Network or server error |
| `2` | Validation error (invalid OAS file, bad input) |
| `3` | Resource not found (API or version) |
| `4` | Conflict (version name already exists) |

## Error messages

Version errors include actionable context:

- **Version not found**: Lists available versions so you can pick the right one
- **Version conflict**: Suggests using `tyk api apply` to update instead
- **API not found**: Suggests using `tyk api list` to find the correct ID
