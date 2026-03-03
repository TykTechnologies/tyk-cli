# Definition of Ready Checklist: OAS Versioning

## US-01: List API Versions

| # | DoR Item | Status | Evidence |
|---|---|---|---|
| 1 | Problem statement clear and in domain language | PASS | "Priya finds it frustrating to SSH into Dashboard UI just to check which versions exist" |
| 2 | User/persona identified with specific characteristics | PASS | Priya Sharma, platform engineer, 12 APIs, terminal-first, CI/CD pipelines |
| 3 | At least 3 domain examples with real data | PASS | Petstore 3 versions, Payments 1 version, non-existent API |
| 4 | UAT scenarios in Given/When/Then (3-7) | PASS | 3 scenarios: multi-version list, JSON output, not-found |
| 5 | Acceptance criteria derived from UAT | PASS | 4 criteria mapped to scenarios |
| 6 | Story right-sized (1-3 days, 3-7 scenarios) | PASS | 1 day estimated, 3 scenarios, replaces existing placeholder |
| 7 | Technical notes identify constraints/dependencies | PASS | Uses existing ListOASAPIVersions client method, replaces placeholder |
| 8 | Dependencies resolved or tracked | PASS | No blocking dependencies, client method exists |

**Result: PASS**

---

## US-02: Create New API Version

| # | DoR Item | Status | Evidence |
|---|---|---|---|
| 1 | Problem statement clear and in domain language | PASS | "6-click Dashboard UI process that is error-prone and not reproducible in CI/CD" |
| 2 | User/persona identified with specific characteristics | PASS | Priya Sharma, publishes new versions regularly, OAS files in git |
| 3 | At least 3 domain examples with real data | PASS | Create v3, create v2 with set-default, conflict v2, invalid OAS |
| 4 | UAT scenarios in Given/When/Then (3-7) | PASS | 5 scenarios: create, set-default, conflict, invalid, JSON |
| 5 | Acceptance criteria derived from UAT | PASS | 6 criteria mapped to scenarios |
| 6 | Story right-sized (1-3 days, 3-7 scenarios) | PASS | 2 days estimated, 5 scenarios |
| 7 | Technical notes identify constraints/dependencies | PASS | Uses CreateOASAPIRequest, conflict detection via list-first |
| 8 | Dependencies resolved or tracked | PASS | Depends on US-01 for conflict detection (tracked) |

**Result: PASS**

---

## US-03: Switch Default API Version

| # | DoR Item | Status | Evidence |
|---|---|---|---|
| 1 | Problem statement clear and in domain language | PASS | "Cannot automate default-version switching in deployment pipeline" |
| 2 | User/persona identified with specific characteristics | PASS | Priya Sharma, deployment runbooks, CI/CD after canary validation |
| 3 | At least 3 domain examples with real data | PASS | Switch v1->v3, already-default no-op, non-existent v99 |
| 4 | UAT scenarios in Given/When/Then (3-7) | PASS | 4 scenarios: switch, no-op, not-found, JSON |
| 5 | Acceptance criteria derived from UAT | PASS | 5 criteria mapped to scenarios |
| 6 | Story right-sized (1-3 days, 3-7 scenarios) | PASS | 1 day estimated, 4 scenarios, uses existing client method |
| 7 | Technical notes identify constraints/dependencies | PASS | Uses SwitchDefaultVersion PATCH, pre-check via ListOASAPIVersions |
| 8 | Dependencies resolved or tracked | PASS | No blocking dependencies, client method exists |

**Result: PASS**

---

## US-04: Declarative Apply with Version Support

| # | DoR Item | Status | Evidence |
|---|---|---|---|
| 1 | Problem statement clear and in domain language | PASS | "Batch apply does not know it is a version -- tries to create duplicate standalone API" |
| 2 | User/persona identified with specific characteristics | PASS | Priya Sharma, GitOps workflow, OAS files in git, CI/CD apply |
| 3 | At least 3 domain examples with real data | PASS | Single file apply, batch with embedded metadata, dry-run |
| 4 | UAT scenarios in Given/When/Then (3-7) | PASS | 4 scenarios: single apply, batch, dry-run, JSON |
| 5 | Acceptance criteria derived from UAT | PASS | 5 criteria mapped to scenarios |
| 6 | Story right-sized (1-3 days, 3-7 scenarios) | PASS | 2-3 days estimated, 4 scenarios |
| 7 | Technical notes identify constraints/dependencies | PASS | Extends classifyContent, new configFileAPIVersion type, ordering |
| 8 | Dependencies resolved or tracked | PASS | Depends on US-02 for version creation logic (tracked) |

**Result: PASS**

---

## US-05: Version-Aware Error Messages

| # | DoR Item | Status | Evidence |
|---|---|---|---|
| 1 | Problem statement clear and in domain language | PASS | "Generic not-found errors force her to switch to Dashboard UI to figure out what went wrong" |
| 2 | User/persona identified with specific characteristics | PASS | Priya Sharma, debugging failed commands, reading CI/CD logs |
| 3 | At least 3 domain examples with real data | PASS | Version not found with hints, conflict with suggestion, API not found |
| 4 | UAT scenarios in Given/When/Then (3-7) | PASS | 3 scenarios: available versions hint, conflict suggestion, list suggestion |
| 5 | Acceptance criteria derived from UAT | PASS | 4 criteria mapped to scenarios |
| 6 | Story right-sized (1-3 days, 3-7 scenarios) | PASS | 1 day estimated, 3 scenarios, cross-cutting helper functions |
| 7 | Technical notes identify constraints/dependencies | PASS | Pattern: fetch available versions on 404, construct suggestion strings |
| 8 | Dependencies resolved or tracked | PASS | Cross-cutting, applies to US-01 through US-04 (tracked) |

**Result: PASS**

---

## Summary

| Story | DoR Status | Estimated Effort | Scenarios |
|---|---|---|---|
| US-01: List API Versions | PASS | 1 day | 3 |
| US-02: Create New API Version | PASS | 2 days | 5 |
| US-03: Switch Default API Version | PASS | 1 day | 4 |
| US-04: Declarative Apply with Version Support | PASS | 2-3 days | 4 |
| US-05: Version-Aware Error Messages | PASS | 1 day | 3 |

**Total estimated effort**: 7-8 days
**Total scenarios**: 19

## Recommended Implementation Order

1. **US-05** (error messages) -- cross-cutting helpers used by all other stories
2. **US-01** (list versions) -- foundation, used by US-02 and US-03 for validation
3. **US-03** (switch default) -- simplest mutation, uses existing client method
4. **US-02** (create version) -- depends on US-01 for conflict detection
5. **US-04** (apply integration) -- depends on US-02 for version creation logic
