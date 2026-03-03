# Acceptance Criteria: OAS Versioning

## US-01: List API Versions

| # | Criterion | UAT Scenario |
|---|---|---|
| AC-01.1 | `tyk api versions list --api-id <id>` displays version names with default marked | List versions for multi-version API |
| AC-01.2 | `--json` outputs `{"api_id", "api_name", "default", "versions"}` to stdout | List versions with JSON output |
| AC-01.3 | Non-existent API returns exit code 3 with "API not found: <id>" | List versions for non-existent API |
| AC-01.4 | Human output shows count summary on stderr | List versions for multi-version API |

## US-02: Create New API Version

| # | Criterion | UAT Scenario |
|---|---|---|
| AC-02.1 | `tyk api versions create --api-id <id> -f <file> --version-name <name>` creates a version linked to base API | Create a new version from OAS file |
| AC-02.2 | `--set-default` flag sets the new version as default on creation | Create version and set as default |
| AC-02.3 | Duplicate version name returns exit code 4 with conflict message and apply suggestion | Reject duplicate version name |
| AC-02.4 | Invalid OAS file returns exit code 2 with validation details | Reject invalid OAS document |
| AC-02.5 | Human output shows switch-default hint when version is not set as default | Create a new version from OAS file |
| AC-02.6 | `--json` outputs `{"action", "api_id", "version_name", "listen_path", "upstream_url", "is_default"}` to stdout | Create version with JSON output |

## US-03: Switch Default API Version

| # | Criterion | UAT Scenario |
|---|---|---|
| AC-03.1 | `tyk api versions switch-default --api-id <id> --version-name <name>` switches default version | Switch default version |
| AC-03.2 | Human output shows previous and new default | Switch default version |
| AC-03.3 | Already-default version is a no-op with exit code 0 | No-op when already default |
| AC-03.4 | Non-existent version returns exit code 3 with available versions list | Version not found |
| AC-03.5 | `--json` outputs `{"action", "api_id", "previous_default", "new_default"}` to stdout | Switch default with JSON output |

## US-04: Declarative Apply with Version Support

| # | Criterion | UAT Scenario |
|---|---|---|
| AC-04.1 | `tyk api apply -f <file> --version-name <name> --base-api-id <id>` creates a version | Apply single file as version |
| AC-04.2 | Batch apply classifies files with `x-tyk-api-gateway.info.versioning` as version files | Batch apply discovers version files |
| AC-04.3 | Version files are applied after their base API in batch ordering | Batch apply discovers version files |
| AC-04.4 | Dry-run reports "would create version" for version files | Dry-run reports version operations |
| AC-04.5 | JSON output includes operation type and version name for version files | Batch apply JSON output |

## US-05: Version-Aware Error Messages

| # | Criterion | UAT Scenario |
|---|---|---|
| AC-05.1 | Version-not-found errors include available versions list | Version not found includes available versions |
| AC-05.2 | Conflict errors include command suggestion for correct action | Conflict error suggests correct command |
| AC-05.3 | API-not-found errors suggest `tyk api list` | API not found suggests list command |
| AC-05.4 | Error messages include API name when available | All version error scenarios |
