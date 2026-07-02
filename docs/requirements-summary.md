# tyk-cli Requirements Summary

This document indexes the requirements catalogued by ReqProof for tyk-cli.
See individual `.req.yaml` files in `specs/` for machine-readable definitions.

<!-- reqproof:req STK-REQ-001 -->
## STK-REQ-001

As an API operator I can manage the full lifecycle of OAS-native APIs — list, retrieve, create, import, update, apply idempotently, and delete — without leaving the terminal

<!-- reqproof:req STK-REQ-002 -->
## STK-REQ-002

As an API operator I can manage access policies — list, retrieve, apply idempotently, delete, and scaffold new policy files — so that I can control which API products clients can consume

<!-- reqproof:req STK-REQ-003 -->
## STK-REQ-003

As an API operator I can configure the CLI's connection to one or more Tyk Dashboard environments — via an interactive wizard, named environments, flags, or environment variables — with clear prec

<!-- reqproof:req SYS-REQ-001 -->
## SYS-REQ-001

api list returns paginated OAS APIs from the Dashboard; page defaults to 1, 10 per page; supports --json and --interactive flags

<!-- reqproof:req SYS-REQ-002 -->
## SYS-REQ-002

api get retrieves a single API by ID; --oas-only strips x-tyk-api-gateway extensions from output; --version-name selects a specific version

<!-- reqproof:req SYS-REQ-003 -->
## SYS-REQ-003

api create builds a minimal OAS document with x-tyk-api-gateway extensions from --name and --upstream-url, auto-generating listen path when --listen-path is omitted

<!-- reqproof:req SYS-REQ-004 -->
## SYS-REQ-004

api import-oas accepts a local file (--file) or remote URL (--url) containing a plain OAS document and creates a new API, always generating a new ID

<!-- reqproof:req SYS-REQ-005 -->
## SYS-REQ-005

'api apply performs an idempotent upsert: if x-tyk-api-gateway.info.id is present it updates the existing API (or creates with that ID if not found); if absent it creates a new API'

<!-- reqproof:req SYS-REQ-006 -->
## SYS-REQ-006

api update-oas updates an existing API's OAS document while preserving its existing x-tyk-api-gateway extensions; the API must already exist or the command exits with code 3

<!-- reqproof:req SYS-REQ-007 -->
## SYS-REQ-007

api delete removes an API by ID; prompts for confirmation unless --yes is passed; exits with code 3 if the API does not exist

<!-- reqproof:req SYS-REQ-008 -->
## SYS-REQ-008

stripExistingAPIID removes x-tyk-api-gateway.info.id from any OAS document before it is sent to the create endpoint, ensuring the Dashboard always generates a new API ID

<!-- reqproof:req SYS-REQ-009 -->
## SYS-REQ-009

AddTykExtensions injects minimal x-tyk-api-gateway block into a plain OAS document; it is a no-op if extensions are already present; it fails if info.title or servers[0].url is missing

<!-- reqproof:req SYS-REQ-010 -->
## SYS-REQ-010

GenerateListenPath converts an API title to a URL-safe slug path (lowercase, non-alphanumeric replaced with hyphens, wrapped in leading/trailing slashes)

<!-- reqproof:req SYS-REQ-011 -->
## SYS-REQ-011

OAS document utilities detect x-tyk-api-gateway presence, extract API ID and upstream URL, and check document structure before transformation

<!-- reqproof:req SYS-REQ-012 -->
## SYS-REQ-012

'API runners (import-oas, update-oas) validate the OAS document structure locally before any Dashboard round-trip: openapi field is non-empty 3.x, info section exists with non-empty title and version.

<!-- reqproof:req SYS-REQ-013 -->
## SYS-REQ-013

All api subcommands exit 0 on success

<!-- reqproof:req SYS-REQ-014 -->
## SYS-REQ-014

All api subcommands exit 2 when required flags are missing or arguments are invalid (before any network call)

<!-- reqproof:req SYS-REQ-015 -->
## SYS-REQ-015

api get and api delete exit 3 when the target API ID is not found (HTTP 404)

<!-- reqproof:req SYS-REQ-016 -->
## SYS-REQ-016

api create, api import-oas, and api apply exit 4 when the Dashboard returns a conflict (HTTP 409)

<!-- reqproof:req SYS-REQ-017 -->
## SYS-REQ-017

API subcommands map authentication failures (HTTP 401) to exit 5, so callers can detect 'rotate the auth token' situations programmatically

<!-- reqproof:req SYS-REQ-018 -->
## SYS-REQ-018

API subcommands map rate-limit responses (HTTP 429) to exit 7, enabling CI scripts to detect throttling and back off

<!-- reqproof:req SYS-REQ-019 -->
## SYS-REQ-019

API subcommands map server errors (HTTP 5xx) to exit 8, distinguishing Dashboard-side failures from local errors

<!-- reqproof:req SYS-REQ-020 -->
## SYS-REQ-020

API subcommands map authorization failures (HTTP 403) to exit 6, distinguishing 'token lacks permission' from authentication failures

<!-- reqproof:req SYS-REQ-021 -->
## SYS-REQ-021

api commands communicate with the Tyk Dashboard via an authenticated HTTP client that sets Authorization and x-org-id headers on every request

<!-- reqproof:req SYS-REQ-022 -->
## SYS-REQ-022

api commands load OAS documents from a local file path (YAML or JSON) or by fetching a URL, returning a normalized map structure

<!-- reqproof:req SYS-REQ-023 -->
## SYS-REQ-023

tyk api versions list/create/switch-default are reserved placeholder subcommands for a future Phase 3 implementation; current invocation emits a fixed placeholder message and returns without contactin

<!-- reqproof:req SYS-REQ-024 -->
## SYS-REQ-024

policy list returns paginated policies from the Dashboard; page defaults to 1; supports --json flag

<!-- reqproof:req SYS-REQ-025 -->
## SYS-REQ-025

policy get retrieves a single policy by ID and outputs CLI-schema YAML to stdout with a summary to stderr; supports --json flag

<!-- reqproof:req SYS-REQ-026 -->
## SYS-REQ-026

'policy apply performs an idempotent upsert: creates a new policy or updates an existing one based on the id field in the YAML file; accepts --file or stdin (-)'

<!-- reqproof:req SYS-REQ-027 -->
## SYS-REQ-027

policy delete removes a policy by ID with an interactive confirmation prompt unless --yes is passed; exits with code 3 if the policy does not exist

<!-- reqproof:req SYS-REQ-028 -->
## SYS-REQ-028

policy init scaffolds a new policy YAML file in CLI schema format with sensible defaults given --id and --name flags

<!-- reqproof:req SYS-REQ-029 -->
## SYS-REQ-029

policy YAML is validated for required fields (friendly ID, access rights) and structural constraints before being submitted to the Dashboard

<!-- reqproof:req SYS-REQ-030 -->
## SYS-REQ-030

policy documents are bi-directionally converted between CLI YAML (PolicyFile) and Dashboard wire JSON (DashboardPolicy) preserving all access rights, duration, and metadata

<!-- reqproof:req SYS-REQ-031 -->
## SYS-REQ-031

ParseDuration parses human-friendly duration strings (e.g. 30d, 1h, 24h) into integer seconds; FormatDuration round-trips seconds back to a human-readable string

<!-- reqproof:req SYS-REQ-032 -->
## SYS-REQ-032

ResolveAccessEntries resolves policy access entries by API name, listen path, or tag against the live API list; returns an error for any selector that matches zero or more than one API

<!-- reqproof:req SYS-REQ-033 -->
## SYS-REQ-033

All policy subcommands exit 0 on success

<!-- reqproof:req SYS-REQ-034 -->
## SYS-REQ-034

All policy subcommands exit 2 on invalid arguments or missing required flags

<!-- reqproof:req SYS-REQ-035 -->
## SYS-REQ-035

policy get and policy delete exit 3 when the target policy ID is not found (HTTP 404)

<!-- reqproof:req SYS-REQ-036 -->
## SYS-REQ-036

policy apply exits 4 on a Dashboard conflict (HTTP 409)

<!-- reqproof:req SYS-REQ-037 -->
## SYS-REQ-037

POL subcommands map authentication failures (HTTP 401) to exit 5, so callers can detect 'rotate the auth token' situations programmatically

<!-- reqproof:req SYS-REQ-038 -->
## SYS-REQ-038

POL subcommands map rate-limit responses (HTTP 429) to exit 7, enabling CI scripts to detect throttling and back off

<!-- reqproof:req SYS-REQ-039 -->
## SYS-REQ-039

POL subcommands map server errors (HTTP 5xx) to exit 8, distinguishing Dashboard-side failures from local errors

<!-- reqproof:req SYS-REQ-040 -->
## SYS-REQ-040

POL subcommands map authorization failures (HTTP 403) to exit 6, distinguishing 'token lacks permission' from authentication failures

<!-- reqproof:req SYS-REQ-041 -->
## SYS-REQ-041

'Configuration is resolved in priority order: CLI flags override environment variables, which override named environments in ~/.config/tyk/cli.toml'

<!-- reqproof:req SYS-REQ-042 -->
## SYS-REQ-042

tyk init runs an interactive wizard that writes one or more named environments to ~/.config/tyk/cli.toml and sets the default environment

<!-- reqproof:req SYS-REQ-043 -->
## SYS-REQ-043

config use switches the active named environment; the switch persists to ~/.config/tyk/cli.toml; config current shows the active environment name and its values

<!-- reqproof:req SYS-REQ-044 -->
## SYS-REQ-044

tyk config list renders every configured environment and marks the active one; emits a friendly empty-state message when no environments are configured

<!-- reqproof:req SYS-REQ-045 -->
## SYS-REQ-045

tyk config set updates dashboard_url, auth_token, and/or org_id on the active environment; refuses with exit 2 when no flags are supplied

<!-- reqproof:req SYS-REQ-046 -->
## SYS-REQ-046

tyk config remove deletes a named environment from cli.toml; refuses to remove an unknown environment, and refuses to remove the sole remaining environment

<!-- reqproof:req SYS-REQ-047 -->
## SYS-REQ-047

Environments may declare a positive timeout_seconds; when set, the HTTP client uses that timeout instead of the 30s default, allowing operators to extend the deadline for slow Dashboards

<!-- reqproof:req SYS-REQ-048 -->
## SYS-REQ-048

Config.Validate rejects any environment missing name, dashboard_url, auth_token, or org_id, and rejects malformed dashboard URLs; returns a descriptive error citing the environment name

<!-- reqproof:req SYS-REQ-049 -->
## SYS-REQ-049

All config subcommands exit 0 on success

<!-- reqproof:req SYS-REQ-050 -->
## SYS-REQ-050

All config subcommands exit 2 on invalid arguments or missing required fields

<!-- reqproof:req SYS-REQ-051 -->
## SYS-REQ-051

config manager loads environments from a TOML file in the XDG base directory, falls back to environment variables, and exposes an effective config resolving precedence

<!-- reqproof:req SYS-REQ-052 -->
## SYS-REQ-052

config manager supports CRUD operations for named environments and default environment selection, persisting changes to the TOML config file

<!-- reqproof:req SW-REQ-001 -->
## SW-REQ-001

FuzzySuggestions returns a slice no longer than the requested max (n argument)

<!-- reqproof:req SW-REQ-002 -->
## SW-REQ-002

FuzzySuggestions returns candidates sorted ascending by Levenshtein distance to the query

<!-- reqproof:req SW-REQ-003 -->
## SW-REQ-003

FuzzySuggestions returns an empty slice when invoked with an empty candidate list

<!-- reqproof:req SW-REQ-004 -->
## SW-REQ-004

ParseDuration returns a non-nil error for empty, negative, fractional, or unsupported-suffix inputs

<!-- reqproof:req SW-REQ-005 -->
## SW-REQ-005

'ParseDuration converts a suffixed input using the canonical multiplier table: s=1, m=60, h=3600, d=86400; the resolved value is positive'

<!-- reqproof:req SW-REQ-006 -->
## SW-REQ-006

'ParseDuration and FormatDuration roundtrip: ParseDuration(FormatDuration(n)) == n for any non-negative n expressible as an exact unit multiple'

<!-- reqproof:req SW-REQ-007 -->
## SW-REQ-007

GenerateListenPath always returns a value that starts with '/' and ends with '/' and is valid as a listen path

<!-- reqproof:req SW-REQ-008 -->
## SW-REQ-008

GenerateListenPath collapses runs of non-alphanumeric characters in the title into a single '-' in the slug

<!-- reqproof:req SW-REQ-009 -->
## SW-REQ-009

GenerateListenPath falls back to '/api/' when the title sanitises to an empty slug, ensuring listen_path_valid even for degenerate input

<!-- reqproof:req INT-REQ-001 -->
## INT-REQ-001

All Dashboard HTTP requests issued by any tyk-cli subcommand shall carry the Authorization and x-org-id headers established by the client.NewClient contract. This interface applies across api, policy,

<!-- reqproof:req INT-REQ-002 -->
## INT-REQ-002

The Dashboard error taxonomy shall be mapped to exit codes via a single typed classifier (classifyDashboardError). Every api and policy subcommand shall route Dashboard errors through the classifier s

<!-- reqproof:req INT-REQ-003 -->
## INT-REQ-003

Every OAS submission path (api import-oas, api update-oas) shall invoke oas.ValidateOASStructure before any Dashboard round-trip. Malformed OAS documents must be rejected locally with ExitBadArgs so t

