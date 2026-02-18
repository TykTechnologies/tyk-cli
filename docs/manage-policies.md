---
title: Manage Policies
nav_order: 4
---

# Manage Policies

Security policies control rate limits, quotas, and API access for your consumers. The CLI lets you author policies as human-friendly YAML files and sync them to your Tyk Dashboard.

## Quick start

```bash
# Scaffold a new policy
tyk policy init --id gold --name "Gold Plan"

# Edit policies/gold.yaml, then apply
tyk policy apply -f policies/gold.yaml

# Verify
tyk policy list
tyk policy get gold
```

## Policy YAML schema

Policies use a **CLI schema** that is converted to Dashboard wire format on apply. Durations are human-readable (`30d`, `1h`, `60s`) and APIs are referenced by name, listen path, or tags instead of raw IDs.

```yaml
id: gold
name: Gold Plan
tags:
  - paid
  - gold
rateLimit:
  requests: 1000
  per: 1m              # 1000 requests per minute
quota:
  limit: 100000
  period: 30d          # 100k requests per 30 days
keyTTL: "0"            # 0 = keys never expire
access:
  - name: users-api    # Match by API name
    versions: [v1]
  - listenPath: /orders/  # Match by listen path
    versions: [v1, v2]
  - tags: [public, v1]    # Match ALL tags, expands to all matching APIs
    versions: [Default]
  - id: abc123def456       # Match by exact API ID
```

### Durations

| Format | Meaning |
|--------|---------|
| `30d` | 30 days |
| `24h` | 24 hours |
| `1m` | 1 minute |
| `60s` | 60 seconds |
| `60` | 60 seconds (plain integer) |
| `0` | Disabled / no expiry |

### API selectors

Each entry in `access` uses exactly **one** selector to identify APIs:

| Selector | Behaviour |
|----------|-----------|
| `name` | Exact match on API name. Must resolve to exactly one API. |
| `listenPath` | Exact match on listen path. Must resolve to exactly one API. |
| `id` | Exact match on API ID. Must resolve to exactly one API. |
| `tags` | Match ALL specified tags. Expands to every API that has all tags. |

If a name or listen path matches zero APIs, the CLI suggests the 3 closest matches.

## Commands

### `tyk policy list`

List policies in a paginated table.

```bash
tyk policy list              # Table with ID, Name, APIs, Tags
tyk policy list --page 2     # Page 2
tyk policy list --json       # JSON to stdout
```

### `tyk policy get <policy-id>`

Retrieve a policy, convert from wire format to CLI schema, and output as YAML.

```bash
tyk policy get gold          # YAML to stdout, summary to stderr
tyk policy get gold --json   # JSON to stdout
```

The CLI reverse-resolves API IDs back to name selectors where possible. If an API ID cannot be resolved (e.g. the API was deleted), it falls back to an `id` selector.

### `tyk policy apply -f <file>`

Apply a policy with **idempotent upsert** semantics: creates the policy if `id` is not found on the server, updates if it exists.

```bash
tyk policy apply -f policies/gold.yaml
tyk policy apply -f -        # Read from stdin
```

The apply pipeline:
1. Parse YAML and validate schema (required fields, duration format, selector format)
2. Resolve selectors against the live API list
3. Parse durations to seconds
4. Convert CLI schema to wire format
5. Create or update the policy

Validation and resolution errors return **exit code 2** with all errors listed so you can fix them in one pass.

### `tyk policy delete <policy-id>`

Delete a policy by ID. Fetches the policy first to show its name in the confirmation prompt.

```bash
tyk policy delete gold       # Interactive confirmation
tyk policy delete gold --yes # Skip confirmation
tyk policy delete gold --yes --json  # JSON result to stdout
```

Exit codes: `0` = success or cancelled, `3` = policy not found.

### `tyk policy init`

Generate a scaffold YAML file with sensible defaults. No Dashboard connectivity required.

```bash
tyk policy init --id gold --name "Gold Plan"
tyk policy init --id silver --name "Silver Plan" --dir ./project
```

Creates `policies/{id}.yaml` relative to `--dir` (default: current directory). Refuses to overwrite an existing file.

## GitOps workflow

Policies are designed for version control. A typical workflow:

```bash
# One-time: scaffold
tyk policy init --id gold --name "Gold Plan"

# Edit in your editor
vim policies/gold.yaml

# Apply to dev
tyk config use dev
tyk policy apply -f policies/gold.yaml

# Commit
git add policies/gold.yaml && git commit -m "Add gold policy"

# Apply to staging
tyk config use staging
tyk policy apply -f policies/gold.yaml
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Network or server error |
| `2` | Validation error (bad input, unresolved selectors) |
| `3` | Resource not found |
