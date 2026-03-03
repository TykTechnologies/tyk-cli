# Acceptance Test Scenarios: OAS Versioning

All scenarios are Given-When-Then specifications that map to Go table-driven tests.
Tests invoke CLI commands through driving ports (`cmd.RunE` or `root.Execute`) with
`httptest` mock servers simulating the Tyk Dashboard API.

Exit code conventions: 0=success, 1=auth failure, 2=validation error, 3=not found, 4=conflict.

---

## US-01: List API Versions

### WS-01: List versions for a multi-version API (Walking Skeleton)

```
Given an API "Petstore" (abc123) exists with versions "v1", "v2", "v3" and default "v1"
When the user runs "tyk api versions list --api-id abc123"
Then stderr contains all 3 version names with "v1" marked as default
And stderr contains "3 versions found"
And the exit code is 0
```

### S-01.2: List versions with JSON output

```
Given an API "Petstore" (abc123) exists with versions "v1", "v2" and default "v1"
When the user runs "tyk api versions list --api-id abc123 --json"
Then stdout contains valid JSON with fields "api_id", "default", "versions"
And "versions" array contains "v1" and "v2"
And "default" equals "v1"
And the exit code is 0
```

### S-01.3: List versions for single-version API

```
Given an API "Payments" (def456) exists with only version "v1" as default
When the user runs "tyk api versions list --api-id def456"
Then stderr contains "1 version found"
And the exit code is 0
```

### S-01.4: List versions for non-existent API

```
Given no API exists with ID "zzz999"
And the Dashboard returns 404 for API "zzz999"
When the user runs "tyk api versions list --api-id zzz999"
Then stderr contains "API not found: zzz999"
And the exit code is 3
```

### S-01.5: List versions with authentication failure

```
Given the Dashboard rejects the auth token with 401
When the user runs "tyk api versions list --api-id abc123"
Then stderr contains "Authentication failed"
And the exit code is 1
```

---

## US-02: Create New API Version

### S-02.1: Create a new version from OAS file

```
Given an API "Petstore" (abc123) exists with version "v1"
And a valid OAS file "petstore-v3.yaml" exists on disk
When the user runs "tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3"
Then the Dashboard receives a POST to /api/apis/oas with base_api_id=abc123 and new_version_name=v3
And stderr contains version "v3" confirmation
And stderr contains a switch-default hint
And the exit code is 0
```

### S-02.2: Create version and set as default

```
Given an API "Petstore" (abc123) exists with default "v1"
And a valid OAS file "petstore-v2.yaml" exists on disk
When the user runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2 --set-default"
Then the Dashboard receives a POST with set_default=true
And stderr confirms version "v2" is the new default
And the exit code is 0
```

### S-02.3: Create version with JSON output

```
Given an API "Petstore" (abc123) exists
And a valid OAS file exists on disk
When the user runs "tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3 --json"
Then stdout contains valid JSON with fields "action", "api_id", "version_name", "is_default"
And "action" equals "created"
And the exit code is 0
```

### S-02.4: Reject duplicate version name

```
Given an API "Petstore" (abc123) exists with version "v2" already present
When the user runs "tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2"
Then stderr contains "already exists"
And stderr contains a suggestion to use "tyk api apply"
And the exit code is 4
```

### S-02.5: Reject invalid OAS document

```
Given a file "broken.yaml" exists that is not a valid OAS document
When the user runs "tyk api versions create --api-id abc123 -f broken.yaml --version-name v2"
Then stderr contains "invalid OAS document"
And the exit code is 2
```

### S-02.6: Create version for non-existent base API

```
Given no API exists with ID "zzz999"
When the user runs "tyk api versions create --api-id zzz999 -f valid.yaml --version-name v2"
Then stderr contains "API not found: zzz999"
And the exit code is 3
```

### S-02.7: Create version with missing required flags

```
When the user runs "tyk api versions create --api-id abc123" without -f or --version-name
Then the command fails with a usage error indicating required flags
And the exit code is non-zero
```

---

## US-03: Switch Default API Version

### S-03.1: Switch default version successfully

```
Given an API "Petstore" (abc123) exists with versions "v1" (default), "v2", "v3"
When the user runs "tyk api versions switch-default --api-id abc123 --version-name v3"
Then the Dashboard receives a PATCH to switch default to "v3"
And stderr shows previous default "v1" and new default "v3"
And the exit code is 0
```

### S-03.2: Switch default with JSON output

```
Given an API "Petstore" (abc123) exists with versions "v1" (default), "v2"
When the user runs "tyk api versions switch-default --api-id abc123 --version-name v2 --json"
Then stdout contains valid JSON with "action", "previous_default", "new_default"
And "previous_default" equals "v1"
And "new_default" equals "v2"
And the exit code is 0
```

### S-03.3: No-op when already the default

```
Given an API "Petstore" (abc123) exists with default "v1"
When the user runs "tyk api versions switch-default --api-id abc123 --version-name v1"
Then no PATCH request is sent to the Dashboard
And stderr contains "already the default"
And the exit code is 0
```

### S-03.4: Version not found with available versions hint

```
Given an API "Petstore" (abc123) exists with versions "v1", "v2"
When the user runs "tyk api versions switch-default --api-id abc123 --version-name v99"
Then stderr contains "not found"
And stderr contains "Available versions" with "v1" and "v2"
And the exit code is 3
```

### S-03.5: Switch default for non-existent API

```
Given no API exists with ID "zzz999"
When the user runs "tyk api versions switch-default --api-id zzz999 --version-name v1"
Then stderr contains "API not found: zzz999"
And the exit code is 3
```

---

## US-04: Declarative Apply with Version Support

### S-04.1: Single file apply as version with flags

```
Given an API "Petstore" (abc123) exists on the Dashboard
And a valid OAS file "petstore-v3.yaml" exists on disk
When the user runs "tyk api apply -f petstore-v3.yaml --version-name v3 --base-api-id abc123"
Then the Dashboard receives a POST creating version "v3" for API abc123
And the exit code is 0
```

### S-04.2: Batch apply discovers version files from metadata

```
Given an API "Petstore" (abc123) exists on the Dashboard
And directory "apis/" contains:
  - "petstore-base.yaml" (standard OAS with x-tyk-api-gateway.info.id: abc123)
  - "petstore-v3.yaml" (OAS with x-tyk-api-gateway.info.versioning.base_api_id: abc123, version_name: v3)
When the user runs "tyk apply -f apis/"
Then the base API is applied first (PUT/POST)
Then the version file is applied second (POST with base_api_id)
And the exit code is 0
```

### S-04.3: Batch apply ordering -- versions after base APIs

```
Given a directory with a policy file, a base API file, and a version file
When the user runs "tyk apply -f dir/"
Then the policy is applied first
Then the base API is applied second
Then the version file is applied third
```

### S-04.4: Dry-run reports version operations accurately

```
Given a valid OAS file "petstore-v3.yaml" exists on disk
When the user runs "tyk api apply -f petstore-v3.yaml --version-name v3 --base-api-id abc123 --dry-run"
Then stderr contains "would create version"
And no mutating requests are sent to the Dashboard
And the exit code is 0
```

### S-04.5: Batch apply JSON output includes version info

```
Given a directory with a version file
When the user runs "tyk apply -f dir/ --json"
Then stdout JSON results include an entry with operation "version_created" and "version_name"
And the exit code is 0
```

### S-04.6: Batch apply with version conflict in directory

```
Given an API "Petstore" (abc123) already has version "v3"
And a directory contains a version file targeting version "v3" of abc123
When the user runs "tyk apply -f dir/"
Then the version file fails with conflict error
And the summary shows 1 failed
```

---

## US-05: Version-Aware Error Messages (Cross-cutting)

These scenarios verify error helper behavior across commands. They overlap with
error scenarios in US-01 through US-04 but focus on message content.

### S-05.1: Version not found includes available versions

```
Given an API "Petstore" (abc123) exists with versions "v1", "v2"
When any version command targets version "v99" on API abc123
Then the error message contains "v99" and lists "v1", "v2" as available
```

### S-05.2: Conflict error suggests correct command

```
Given version "v2" already exists for API abc123
When the user runs "tyk api versions create" targeting version "v2"
Then the error message contains "tyk api apply"
```

### S-05.3: API not found suggests list command

```
Given no API exists with ID "nonexistent"
When any version command targets API "nonexistent"
Then the error message contains "tyk api list"
```

### S-05.4: Error messages include API name when available

```
Given an API named "Petstore" (abc123) exists but version "v99" does not
When the user runs switch-default for "v99" on abc123
Then the error message contains "Petstore" (not just the ID)
```

---

## Coverage Summary

| Story | Total Scenarios | Happy Path | Error/Edge | Error Ratio |
|-------|----------------|------------|------------|-------------|
| US-01 | 5 | 3 | 2 | 40% |
| US-02 | 7 | 3 | 4 | 57% |
| US-03 | 5 | 2 | 3 | 60% |
| US-04 | 6 | 4 | 2 | 33% |
| US-05 | 4 | 0 | 4 | 100% |
| **Total** | **27** | **12** | **15** | **56%** |

Error path ratio: 15/27 = 56% (exceeds 40% target).
