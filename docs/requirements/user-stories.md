# User Stories: OAS Versioning

---

# US-01: List API Versions

## Problem (The Pain)
Priya Sharma is a platform engineer managing 12 APIs on a shared Tyk gateway.
She finds it frustrating to SSH into the Dashboard UI just to check which versions
exist for an API and which one is currently serving default traffic.

## Who (The User)
- Platform engineer managing multiple APIs across environments
- Works primarily in terminal, uses CLI in CI/CD pipelines
- Needs quick orientation before making changes

## Solution (What We Build)
A `tyk api versions list` command that shows all versions for an API with the
default clearly marked, supporting both human and JSON output.

## Domain Examples

### Example 1: Petstore with three versions
Priya runs `tyk api versions list --api-id abc123` for her Petstore API.
She sees versions v1 (marked as default), v2, and v3 in a clean table.
The summary line reads "3 versions found. Default: v1".

### Example 2: API with single version
Priya runs `tyk api versions list --api-id def456` for her Payments API.
She sees only "v1" marked as default. Summary: "1 version found. Default: v1".

### Example 3: Non-existent API
Priya mistypes the ID: `tyk api versions list --api-id zzz999`.
She gets "Error: API not found: zzz999" on stderr, exit code 3.

## UAT Scenarios (BDD)

### Scenario: List versions for multi-version API
Given Priya has an API "Petstore" (abc123) with versions "v1", "v2", "v3" and default "v1"
When Priya runs "tyk api versions list --api-id abc123"
Then the output shows all 3 versions with "v1" marked as default
And the exit code is 0

### Scenario: List versions with JSON output
Given Priya has an API "Petstore" (abc123) with versions "v1", "v2" and default "v1"
When Priya runs "tyk api versions list --api-id abc123 --json"
Then stdout contains JSON with "api_id", "default", and "versions" array
And the exit code is 0

### Scenario: List versions for non-existent API
Given no API exists with ID "zzz999"
When Priya runs "tyk api versions list --api-id zzz999"
Then stderr shows "Error: API not found: zzz999"
And the exit code is 3

## Acceptance Criteria
- [ ] Command `tyk api versions list --api-id <id>` returns version names and default marker
- [ ] `--json` flag outputs structured JSON to stdout with api_id, default, versions
- [ ] Non-existent API returns exit code 3 with descriptive error
- [ ] Human output shows version count summary on stderr

## Technical Notes
- Uses existing `client.ListOASAPIVersions(ctx, apiID)` which returns `([]string, string, error)`
- Replace placeholder in `NewAPIVersionsListCommand()` at `internal/cli/api.go`
- Exit code mapping: 0=success, 3=not-found (from ErrorResponse.Status 404)

---

# US-02: Create New API Version

## Problem (The Pain)
Priya Sharma has a new version of her Petstore API spec (OpenAPI 3.0) and needs to
publish it as version "v3" on the gateway. Currently she has to use the Dashboard UI
to upload the spec and link it to the base API, a 6-click process that is error-prone
and not reproducible in CI/CD.

## Who (The User)
- Platform engineer publishing new API versions regularly
- Needs reproducible, scriptable version creation
- Works with OAS files stored in git repositories

## Solution (What We Build)
A `tyk api versions create` command that takes an OAS file and creates a new version
linked to an existing base API, with optional default-setting.

## Domain Examples

### Example 1: Create v3 of Petstore
Priya runs `tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3`.
The version is created. Output shows: API name, version "v3", listen path "/petstore-v3/",
upstream "https://api.petstore.io/v3", default "no". A hint suggests the switch-default command.

### Example 2: Create v2 and set as default
Priya runs `tyk api versions create --api-id def456 -f payments-v2.yaml --version-name v2 --set-default`.
Version "v2" is created and immediately set as the default. Output shows default "yes".

### Example 3: Conflicting version name
Priya accidentally runs create for "v2" when v2 already exists on abc123.
She gets "Error: version \"v2\" already exists for API abc123" with exit code 4
and a hint to use `tyk api apply` for updates.

### Example 4: Invalid OAS file
Priya points to a malformed YAML file missing the info section.
She gets "Error: invalid OAS document: missing info section" with exit code 2.

## UAT Scenarios (BDD)

### Scenario: Create a new version from OAS file
Given Priya has an API "Petstore" (abc123) with version "v1"
And Priya has a valid OAS file "petstore-v3.yaml" with upstream "https://api.petstore.io/v3"
When Priya runs "tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3"
Then version "v3" is created for API abc123
And stderr shows version details and switch-default hint
And the exit code is 0

### Scenario: Create version and set as default
Given Priya has an API "Petstore" (abc123) with default "v1"
When Priya runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2 --set-default"
Then version "v2" is created and set as default
And the exit code is 0

### Scenario: Reject duplicate version name
Given Priya has an API "Petstore" (abc123) with version "v2" already existing
When Priya runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2"
Then stderr shows conflict error with suggestion to use apply
And the exit code is 4

### Scenario: Reject invalid OAS document
Given Priya has a file "broken.yaml" without an info section
When Priya runs "tyk api versions create --api-id abc123 -f broken.yaml --version-name v2"
Then stderr shows validation error
And the exit code is 2

### Scenario: Create version with JSON output
Given Priya has an API "Petstore" (abc123)
When Priya runs "tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3 --json"
Then stdout contains JSON with action, api_id, version_name, listen_path, is_default
And the exit code is 0

## Acceptance Criteria
- [ ] Command creates a version linked to base API via `CreateOASAPIRequest.BaseAPIID`
- [ ] `--set-default` flag sets the new version as default
- [ ] Duplicate version name returns exit code 4 with conflict message
- [ ] Invalid OAS file returns exit code 2 with validation details
- [ ] Human output includes switch-default hint when not set as default
- [ ] `--json` outputs structured result to stdout

## Technical Notes
- Uses `CreateOASAPIRequest` with `BaseAPIID`, `NewVersionName`, `SetDefault` fields
- OAS file loading reuses existing `loadOASFile()` pattern from api.go
- Version conflict detection: list versions first, check for name collision before POST
- Depends on: US-01 (list versions for conflict detection)

---

# US-03: Switch Default API Version

## Problem (The Pain)
Priya Sharma has deployed version "v3" of her Petstore API and validated it with
canary traffic. She needs to switch default traffic from "v1" to "v3". Currently
this requires navigating the Dashboard UI, finding the right API, and clicking
through a confirmation dialog -- a process she cannot automate in her deployment pipeline.

## Who (The User)
- Platform engineer executing deployment runbooks
- Needs atomic, scriptable default-version switching
- Often runs this as part of CI/CD after canary validation

## Solution (What We Build)
A `tyk api versions switch-default` command that atomically switches which version
serves as the default, showing the transition clearly.

## Domain Examples

### Example 1: Switch from v1 to v3
Priya runs `tyk api versions switch-default --api-id abc123 --version-name v3`.
Output: "Previous default: v1, New default: v3".

### Example 2: Already the default (no-op)
Priya runs the command for v1 when v1 is already default.
Output: "Version \"v1\" is already the default for API abc123." Exit code 0.

### Example 3: Non-existent version
Priya runs switch-default for "v99" which does not exist.
Output: "Error: version \"v99\" not found for API abc123. Available versions: v1, v2, v3."
Exit code 3.

## UAT Scenarios (BDD)

### Scenario: Switch default version
Given Priya has an API "Petstore" (abc123) with versions "v1" (default), "v2", "v3"
When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v3"
Then the default is switched to "v3"
And stderr shows previous default "v1" and new default "v3"
And the exit code is 0

### Scenario: No-op when already default
Given Priya has an API "Petstore" (abc123) with default "v1"
When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v1"
Then stderr shows already-default message
And the exit code is 0

### Scenario: Version not found
Given Priya has an API "Petstore" (abc123) with versions "v1", "v2"
When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v99"
Then stderr shows not-found error with available versions
And the exit code is 3

### Scenario: Switch default with JSON output
Given Priya has an API "Petstore" (abc123) with versions "v1" (default), "v2"
When Priya runs "tyk api versions switch-default --api-id abc123 --version-name v2 --json"
Then stdout contains JSON with action, previous_default, new_default
And the exit code is 0

## Acceptance Criteria
- [ ] Command switches default version via `client.SwitchDefaultVersion()`
- [ ] Shows previous and new default in human output
- [ ] No-op with exit 0 when target is already default
- [ ] Version not found returns exit code 3 with available versions list
- [ ] `--json` outputs structured result to stdout

## Technical Notes
- Uses existing `client.SwitchDefaultVersion(ctx, apiID, versionName)` -- PATCH endpoint
- Pre-check: call `ListOASAPIVersions` to get current default and validate target exists
- No-op detection: compare requested version with current default before calling PATCH

---

# US-04: Declarative Apply with Version Support

## Problem (The Pain)
Priya Sharma manages her API configurations in a git repository. She uses `tyk apply -f apis/`
to sync configurations to the gateway. When she adds a new version file for an existing API,
the batch apply does not know it is a version -- it tries to create a new standalone API,
which fails or creates a duplicate.

## Who (The User)
- Platform engineer using GitOps workflow for API management
- Stores OAS files in git, applies them via CI/CD
- Needs version relationships to be declarative in the files themselves

## Solution (What We Build)
Extend `tyk api apply` and `tyk apply -f` to recognize version metadata in OAS files
and create/update versions accordingly, using a `versioning` section in the Tyk extension.

## Domain Examples

### Example 1: Single file apply as version
Priya runs `tyk api apply -f petstore-v3.yaml --version-name v3 --base-api-id abc123`.
The file is applied as version "v3" of the Petstore API, not as a new standalone API.

### Example 2: Batch apply with embedded version metadata
Priya has a directory with `petstore-base.yaml` and `petstore-v3.yaml`. The v3 file
contains `x-tyk-api-gateway.info.versioning.base_api_id: abc123` and `version_name: v3`.
Running `tyk apply -f apis/` updates the base and creates v3 as a version.

### Example 3: Dry-run shows version operations
Priya runs `tyk apply -f apis/ --dry-run`. Output shows:
"[1/2] petstore-base.yaml ... would update"
"[2/2] petstore-v3.yaml ... would create version: v3 of abc123"

## UAT Scenarios (BDD)

### Scenario: Apply single file as version with flags
Given Priya has an API "Petstore" (abc123)
And a valid OAS file "petstore-v3.yaml"
When Priya runs "tyk api apply -f petstore-v3.yaml --version-name v3 --base-api-id abc123"
Then version "v3" is created for API abc123
And the exit code is 0

### Scenario: Batch apply discovers version files
Given Priya has an API "Petstore" (abc123)
And directory "apis/" contains a base file and a version file with embedded versioning metadata
When Priya runs "tyk apply -f apis/"
Then the base API is updated and the version is created
And the exit code is 0

### Scenario: Dry-run reports version operations
Given Priya has an API "Petstore" (abc123)
And a valid OAS file "petstore-v3.yaml"
When Priya runs "tyk api apply -f petstore-v3.yaml --version-name v3 --base-api-id abc123 --dry-run"
Then stderr shows "would create version" without making changes
And the exit code is 0

### Scenario: Batch apply JSON output includes version info
Given version files in a directory
When Priya runs "tyk apply -f apis/ --json"
Then stdout JSON results include operation "version_created" for version files
And the exit code is 0

## Acceptance Criteria
- [ ] `tyk api apply` accepts `--version-name` and `--base-api-id` flags for version creation
- [ ] Batch apply classifies files with `x-tyk-api-gateway.info.versioning` as version files
- [ ] Version files are applied after their base API (ordering: policies, base APIs, version files)
- [ ] Dry-run accurately reports "would create version" for version files
- [ ] JSON output includes version operation type and version name

## Technical Notes
- Extend `classifyContent()` in apply.go to detect versioning metadata
- Add new `configFileType`: `configFileAPIVersion`
- Extend `sortFiles()`: policies first, then base APIs, then version files
- Version files use `CreateOASAPIRequest` with `BaseAPIID` and `NewVersionName`
- Depends on: US-02 (create version logic)

---

# US-05: Version-Aware Error Messages

## Problem (The Pain)
Priya Sharma runs a version command that fails -- the API does not exist, or the version
name is wrong. The current placeholder commands give no useful feedback. Even once
implemented, generic "not found" errors force her to switch to the Dashboard UI to
figure out what went wrong.

## Who (The User)
- Platform engineer debugging failed CLI commands
- Needs actionable error context without leaving the terminal
- May be reading errors in CI/CD logs after the fact

## Solution (What We Build)
Error messages for version operations that include actionable context: available versions
when a version is not found, the correct command when a conflict occurs, and the API name
alongside the ID for orientation.

## Domain Examples

### Example 1: Version not found with hints
Priya runs `tyk api get --api-id abc123 --version-name v99`.
Error: "version \"v99\" not found for API abc123 (Petstore). Available versions: v1, v2, v3."

### Example 2: Version conflict with command suggestion
Priya runs `tyk api versions create --api-id abc123 --version-name v2` but v2 exists.
Error: "version \"v2\" already exists for API abc123. Use 'tyk api apply -f <file>' to update."

### Example 3: API not found on version operation
Priya runs `tyk api versions list --api-id nonexistent`.
Error: "API not found: nonexistent. Use 'tyk api list' to see available APIs."

## UAT Scenarios (BDD)

### Scenario: Version not found includes available versions
Given Priya has an API "Petstore" (abc123) with versions "v1", "v2"
When Priya runs any version command targeting "v99"
Then the error message lists available versions "v1", "v2"

### Scenario: Conflict error suggests correct command
Given version "v2" already exists for API abc123
When Priya runs "tyk api versions create" with version-name "v2"
Then the error suggests using "tyk api apply" instead

### Scenario: API not found suggests list command
Given no API exists with ID "nonexistent"
When Priya runs any version command for "nonexistent"
Then the error suggests using "tyk api list"

## Acceptance Criteria
- [ ] Version-not-found errors include the list of available versions
- [ ] Conflict errors include a command suggestion for the correct action
- [ ] API-not-found errors include a suggestion to use `tyk api list`
- [ ] All error messages include the API name when available (not just the ID)

## Technical Notes
- Pattern: on 404, call `ListOASAPIVersions` to populate available versions in error message
- Pattern: on 409/conflict, construct suggestion string with correct command
- API name lookup: call `GetOASAPI` on 404 before returning error (if API exists but version does not)
- Cross-cutting concern: applies to US-01, US-02, US-03, US-04
