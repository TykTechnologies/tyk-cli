# Data Models: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DESIGN
**Date**: 2026-02-18

---

## 1. CLI-Side Policy Types (What Users Write in YAML)

### PolicyFile -- Top-Level YAML Structure

```yaml
apiVersion: tyk.tyktech/v1        # Required, always "tyk.tyktech/v1"
kind: Policy                       # Required, always "Policy"
metadata:
  id: gold                         # Required, unique per org
  name: Gold Plan                  # Required, human-readable
  tags: [gold, paid]               # Optional, string array
spec:
  rateLimit:
    requests: 1000                 # Optional, integer (0 = unlimited)
    per: 60                        # Optional, duration string or seconds
  quota:
    limit: 100000                  # Optional, integer (0 = unlimited)
    period: 30d                    # Optional, duration string or seconds
  keyTTL: 0                        # Optional, duration or 0 (never expires)
  access:                          # Required, at least 1 entry
    - name: users-api              # Exactly one of: id | name | listenPath | tags
      versions: [v1]               # Optional, defaults to API's default version
    - listenPath: /orders/
      versions: [v1, v2]
    - tags: [public, v1]
      versions: [v1]
    - id: foobar123
```

### Go Struct Mapping

The crafter defines the exact Go struct tags, field visibility, and package organization. The following describes the semantic contract:

| YAML Path | Go Semantic | Type | Required | Validation |
|---|---|---|---|---|
| `apiVersion` | API version string | string | Yes | Must equal `"tyk.tyktech/v1"` |
| `kind` | Resource kind | string | Yes | Must equal `"Policy"` |
| `metadata.id` | Policy identifier | string | Yes | Non-empty |
| `metadata.name` | Display name | string | Yes | Non-empty |
| `metadata.tags` | Classification tags | []string | No | - |
| `spec.rateLimit.requests` | Max requests | int64 | No | >= 0 |
| `spec.rateLimit.per` | Rate window | duration/int64 | No | Valid duration or positive int |
| `spec.quota.limit` | Max quota | int64 | No | >= 0 |
| `spec.quota.period` | Quota window | duration/int64 | No | Valid duration or positive int |
| `spec.keyTTL` | Key expiry | duration/int64 | No | Valid duration or >= 0 |
| `spec.access` | API access list | []AccessEntry | Yes | At least 1 entry |
| `spec.access[].id` | API ID selector | string | One of four | - |
| `spec.access[].name` | API name selector | string | One of four | - |
| `spec.access[].listenPath` | Listen path selector | string | One of four | - |
| `spec.access[].tags` | Tag-based selector | []string | One of four | At least 1 tag |
| `spec.access[].versions` | API version names | []string | No | Defaults to API's default |

### Selector Constraint

Each access entry must set **exactly one** of `id`, `name`, `listenPath`, or `tags`. Setting zero or multiple is a validation error.

### Duration Format

Duration fields accept either:
- Plain integer: interpreted as seconds (e.g., `60`)
- Duration string: integer + suffix `s`/`m`/`h`/`d` (e.g., `"1m"`, `"30d"`)

The YAML parser sees plain integers as int and suffixed strings as string. The types layer should handle both representations (the crafter decides how -- e.g., a custom unmarshal method or a union type).

---

## 2. Wire Types (Dashboard API Format)

### DashboardPolicy -- What the Dashboard Returns/Accepts

Based on Tyk Dashboard REST API policy schema:

```json
{
  "_id": "gold",
  "id": "",
  "name": "Gold Plan",
  "org_id": "5e9d9544a1dcd60001d0ed20",
  "rate": 1000,
  "per": 60,
  "quota_max": 100000,
  "quota_renewal_rate": 2592000,
  "key_expires_in": 0,
  "tags": ["gold", "paid"],
  "access_rights": {
    "a1b2c3d4e5f6": {
      "api_id": "a1b2c3d4e5f6",
      "api_name": "users-api",
      "versions": ["v1"],
      "allowed_urls": [],
      "limit": null
    },
    "g7h8i9j0k1l2": {
      "api_id": "g7h8i9j0k1l2",
      "api_name": "orders-api",
      "versions": ["v1", "v2"],
      "allowed_urls": [],
      "limit": null
    }
  },
  "active": true,
  "is_inactive": false
}
```

### Wire Field Mapping

| Dashboard Wire Field | Go Semantic | Type | Notes |
|---|---|---|---|
| `_id` | Policy ID (MongoDB ID or custom string) | string | Used for GET/PUT/DELETE path param |
| `id` | Internal numeric ID | string | Often empty; `_id` is the primary identifier |
| `name` | Policy name | string | - |
| `org_id` | Organization ID | string | Set from CLI config |
| `rate` | Rate limit requests | int64 | 0 = unlimited |
| `per` | Rate limit window (seconds) | int64 | - |
| `quota_max` | Quota limit | int64 | -1 = unlimited |
| `quota_renewal_rate` | Quota period (seconds) | int64 | - |
| `key_expires_in` | Key TTL (seconds) | int64 | 0 = never expires |
| `tags` | Tags array | []string | - |
| `access_rights` | API access map | map[string]AccessRight | Keyed by API ID |
| `active` | Policy active flag | bool | Default true |
| `is_inactive` | Inverse of active | bool | Default false |

### AccessRight (Wire Format, Per API)

| Field | Type | Notes |
|---|---|---|
| `api_id` | string | Dashboard API ID |
| `api_name` | string | API name (informational) |
| `versions` | []string | Allowed version names |
| `allowed_urls` | []AllowedURL | URL-level restrictions (Phase 1: empty) |
| `limit` | *RateQuotaLimit | Per-API limits (Phase 1: null) |

### Dashboard Policy List Response

```json
{
  "Data": [
    { "_id": "gold", "name": "Gold Plan", ... },
    { "_id": "silver", "name": "Silver Plan", ... }
  ],
  "Pages": 1,
  "StatusCode": 200
}
```

| Field | Type | Notes |
|---|---|---|
| `Data` | []DashboardPolicy | Array of policy objects |
| `Pages` | int | Total number of pages |
| `StatusCode` | int | HTTP status code |

---

## 3. Conversion Functions

### CLI -> Wire (for `apply`)

| CLI Field | Wire Field | Conversion |
|---|---|---|
| `metadata.id` | `_id` | Direct copy |
| `metadata.name` | `name` | Direct copy |
| `metadata.tags` | `tags` | Direct copy |
| `spec.rateLimit.requests` | `rate` | Direct copy (already int) |
| `spec.rateLimit.per` | `per` | `ParseDuration` -> seconds |
| `spec.quota.limit` | `quota_max` | Direct copy (already int) |
| `spec.quota.period` | `quota_renewal_rate` | `ParseDuration` -> seconds |
| `spec.keyTTL` | `key_expires_in` | `ParseDuration` -> seconds |
| `spec.access` | `access_rights` | See below |
| (implicit) | `org_id` | From CLI config |
| (implicit) | `active` | `true` |
| (implicit) | `is_inactive` | `false` |

#### Access Entry Conversion (CLI -> Wire)

```
For each AccessEntry in spec.access:
  1. Resolve selector to API ID(s):
     - id: use directly
     - name: resolve via API list
     - listenPath: resolve via API list
     - tags: resolve via API list (may expand to multiple APIs)
  2. For each resolved API ID:
     access_rights[apiID] = {
       api_id: apiID,
       api_name: resolved_api_name,
       versions: entry.versions or [api_default_version],
       allowed_urls: [],
       limit: null
     }
```

### Wire -> CLI (for `get`)

| Wire Field | CLI Field | Conversion |
|---|---|---|
| `_id` | `metadata.id` | Direct copy |
| `name` | `metadata.name` | Direct copy |
| `tags` | `metadata.tags` | Direct copy |
| `rate` | `spec.rateLimit.requests` | Direct copy |
| `per` | `spec.rateLimit.per` | `FormatDuration` -> best human unit |
| `quota_max` | `spec.quota.limit` | Direct copy |
| `quota_renewal_rate` | `spec.quota.period` | `FormatDuration` -> best human unit |
| `key_expires_in` | `spec.keyTTL` | `FormatDuration` -> best human unit |
| `access_rights` | `spec.access` | See below |

#### Access Rights Reverse Conversion (Wire -> CLI)

```
For each (apiID, accessRight) in access_rights:
  1. Best-effort reverse resolution via API list:
     - If apiID matches a known API: use name as selector
     - If API not found: fall back to id selector
  2. Build AccessEntry:
     - name: api_name (or id: apiID if not resolved)
     - versions: accessRight.versions
```

The reverse resolution is best-effort. If an API has been deleted since the policy was created, the export uses the `id` selector as fallback. This is safe and re-applicable.

---

## 4. Selector Types and Resolution Results

### Selector Input (from YAML)

Each access entry contains exactly one selector. The selector type is determined by which field is set:

| Field Set | Selector Type | Resolution Behavior |
|---|---|---|
| `id` | Direct ID | Must match exactly 1 API by ID |
| `name` | Name match | Must match exactly 1 API by name |
| `listenPath` | Path match | Must match exactly 1 API by listen path |
| `tags` | Tag intersection | Must match >= 1 API having ALL listed tags |

### Resolution Result (Per Access Entry)

Each entry resolves to one of:

- **Success**: one or more resolved API IDs with their metadata
- **NotFound**: zero matches with fuzzy suggestions (top 3 closest by edit distance)
- **Ambiguous**: multiple matches for a uniqueness-required selector (name/listenPath/id) with candidate list

### Resolution Result Aggregate

The resolver processes ALL access entries and returns:

- **All resolved**: map of entry index -> resolved API ID(s)
- **Any errors**: list of resolution errors with entry index, selector value, and error details

All errors are collected before returning. The CLI layer displays all errors together, not one at a time.

### Fuzzy Suggestion Data

For "Did you mean?" suggestions:

| Field | Type | Description |
|---|---|---|
| `suggested_name` | string | API name or path closest to the selector |
| `suggested_id` | string | API ID for the suggestion |
| `distance` | int | Edit distance (lower = closer match) |

Top 3 suggestions are returned, sorted by distance ascending.

---

## 5. Duration Type Representation

### Parse Input/Output

| Input | Parsed Output | Formatted Output |
|---|---|---|
| `"30d"` | `2592000` (int64 seconds) | `"30d"` |
| `"24h"` | `86400` | `"24h"` or `"1d"` (prefer largest clean unit) |
| `"1m"` | `60` | `"1m"` |
| `"60s"` | `60` | `"1m"` |
| `60` | `60` | `"1m"` |
| `0` | `0` | `0` |

### Formatting Rules (Wire -> CLI)

`FormatDuration` picks the largest unit that divides evenly:
1. If value == 0: return `0`
2. If value % 86400 == 0: return `"{value/86400}d"`
3. If value % 3600 == 0: return `"{value/3600}h"`
4. If value % 60 == 0: return `"{value/60}m"`
5. Else: return `"{value}s"`

---

## 6. Validation Error Types

### Structure

| Field | Type | Description |
|---|---|---|
| `field` | string | Dot-path to field (e.g., `spec.access[0].name`) |
| `message` | string | Human-readable error description |
| `kind` | string | Error category: `schema`, `duration`, `selector` |

### Example Errors

```
field: metadata.id     kind: schema     message: "required field missing"
field: spec.rateLimit.per  kind: duration   message: "invalid duration 'abc': expected integer or NNs/NNm/NNh/NNd"
field: spec.access[0]  kind: selector   message: "exactly one of id, name, listenPath, or tags must be set"
field: spec.access[1].name  kind: selector  message: "no API found for name 'inventori-api'. Did you mean: inventory-api (m3n4o5p6q7r8)?"
```
