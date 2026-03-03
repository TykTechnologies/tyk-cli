# Shared Artifacts Registry: OAS Versioning

## Artifacts

| Artifact | Type | Source | Consumed By | Notes |
|---|---|---|---|---|
| `api_id` | string | Dashboard API (assigned on create, or from `x-tyk-api-gateway.info.id`) | list-versions, get-version, create-version, switch-default, apply | Primary key for all version operations |
| `version_name` | string | User flag `--version-name`, or extracted from OAS `info.version`, or from `x-tyk-api-gateway.info.versioning.version_name` | get-version, create-version, switch-default, apply | Must be unique within an API |
| `default_version` | string | `OASAPI.DefaultVersion` from Dashboard response | list-versions (display), switch-default (read previous, write new) | Exactly one version is default at any time |
| `oas_document` | map | Local file (`-f` flag) or Dashboard GET response | create-version (input), get-version (output), apply (input) | Full OAS 3.x document with optional Tyk extensions |
| `version_list` | []string | `VersionListResponse.Versions` from `GET /api/apis/oas/{id}/versions` | list-versions (output), error messages (available versions hint) | Used to validate version existence and show hints |
| `listen_path` | string | Extracted from OAS `x-tyk-api-gateway.server.listenPath.value` or `APIVersion.ListenPath` | create-version (output), get-version (output) | Per-version, may differ from base API |
| `upstream_url` | string | Extracted from OAS `x-tyk-api-gateway.upstream.url` or `APIVersion.UpstreamURL` | create-version (output), get-version (output) | Per-version, may differ from base API |
| `base_api_id` | string | User flag `--base-api-id` or from `x-tyk-api-gateway.info.versioning.base_api_id` in OAS file | apply (to link version to parent API) | Only needed when creating versions via apply |

## Cross-Step Data Flow

```
list-versions ──> version_list, default_version
                      |
                      v
get-version <── version_name (from user or from list)
     |
     v
  oas_document ──> (inspect content)
                      |
                      v
create-version <── oas_document (from file), version_name, api_id
     |
     v
  version_name ──> switch-default
     |
     v
  default_version (updated)
```

## Versioning Extension Convention

For batch apply, version files declare their relationship to a base API:

```yaml
x-tyk-api-gateway:
  info:
    name: "Petstore"
    versioning:
      base_api_id: "abc123"        # links to parent API
      version_name: "v3"           # version identifier
      set_default: false           # whether to make this the default
```

This extension is consumed by `tyk apply -f` when processing directories. It allows the batch applier to distinguish between "update the base API" and "create/update a version of the base API."

## Exit Code Mapping

| Code | Meaning | Triggered By |
|---|---|---|
| 0 | Success | All happy paths, already-default no-op |
| 1 | General error | Auth failure, network error |
| 2 | Validation error | Invalid OAS document, missing required flags |
| 3 | Not found | API not found, version not found |
| 4 | Conflict | Version name already exists |
