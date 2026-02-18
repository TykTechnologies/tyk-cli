# Test Scenario Inventory: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DISTILL
**Date**: 2026-02-18

---

## Summary

| Category | Count |
|---|---|
| Walking skeleton scenarios | 4 |
| Focused happy-path scenarios | 24 |
| Error/edge-case scenarios | 27 |
| **Total scenarios** | **55** |
| Error path ratio | **49%** (27/55) -- exceeds 40% target |

---

## Story-to-Scenario Mapping

### US-PM-01: List Security Policies

| # | Scenario | Type | Feature File | Go Test Function |
|---|---|---|---|---|
| 1 | List policies empty inventory | Walking skeleton | walking-skeleton.feature | TestPolicyList_Empty |
| 2 | List policies with existing data | Walking skeleton | walking-skeleton.feature | TestPolicyList_WithPolicies |
| 3 | List policies in JSON format | Happy path | milestone-1-list-get.feature | TestPolicyList_JSONOutput |
| 4 | List policies with pagination | Happy path | milestone-1-list-get.feature | TestPolicyList_Pagination_EmptyPage |
| 5 | List policies shows API count | Happy path | milestone-1-list-get.feature | TestPolicyList_APICount |
| 6 | List policies displays tags | Happy path | milestone-1-list-get.feature | TestPolicyList_Tags |

### US-PM-02: Get Policy Details

| # | Scenario | Type | Feature File | Go Test Function |
|---|---|---|---|---|
| 7 | Get policy human-readable | Happy path | milestone-1-list-get.feature | TestPolicyGet_Human |
| 8 | Get policy exports CLI schema | Happy path | milestone-1-list-get.feature | TestPolicyGet_CLISchema |
| 9 | Get policy JSON format | Happy path | milestone-1-list-get.feature | TestPolicyGet_JSON |
| 10 | Get non-existent policy | Error path | milestone-1-list-get.feature | TestPolicyGet_NotFound |
| 11 | Get policy redirect to file | Happy path | milestone-1-list-get.feature | TestPolicyGet_FileExport |

### US-PM-03: Apply Policy from File

| # | Scenario | Type | Feature File | Go Test Function |
|---|---|---|---|---|
| 12 | Apply new policy name selector | Walking skeleton | walking-skeleton.feature | TestPolicyApply_Create_NameSelector |
| 13 | Apply update existing policy | Walking skeleton | walking-skeleton.feature | TestPolicyApply_Update_Idempotent |
| 14 | Apply listenPath selector | Happy path | milestone-2-apply.feature | TestPolicyApply_ListenPathSelector |
| 15 | Apply direct ID selector | Happy path | milestone-2-apply.feature | TestPolicyApply_IDSelector |
| 16 | Apply tags selector multi-match | Happy path | milestone-2-apply.feature | TestPolicyApply_TagsSelector |
| 17 | Apply multiple access entries | Happy path | milestone-2-apply.feature | TestPolicyApply_MultipleSelectors |
| 18 | Apply duration conversion | Happy path | milestone-2-apply.feature | TestPolicyApply_DurationConversion |
| 19 | Apply integer seconds | Happy path | milestone-2-apply.feature | TestPolicyApply_IntegerSeconds |
| 20 | Apply keyTTL zero | Happy path | milestone-2-apply.feature | TestPolicyApply_KeyTTLZero |
| 21 | Apply from stdin | Happy path | milestone-2-apply.feature | TestPolicyApply_Stdin |
| 22 | Apply JSON output | Happy path | milestone-2-apply.feature | TestPolicyApply_JSONOutput |
| 23 | Apply name matches zero APIs | Error path | milestone-2-apply.feature | TestPolicyApply_NameNotFound |
| 24 | Apply name matches multiple | Error path | milestone-2-apply.feature | TestPolicyApply_NameAmbiguous |
| 25 | Apply listenPath ambiguous | Error path | milestone-2-apply.feature | TestPolicyApply_ListenPathAmbiguous |
| 26 | Apply tags match zero | Error path | milestone-2-apply.feature | TestPolicyApply_TagsEmpty |
| 27 | Apply ID not found | Error path | milestone-2-apply.feature | TestPolicyApply_IDNotFound |
| 28 | Apply missing metadata.id | Error path | milestone-2-apply.feature | TestPolicyApply_MissingID |
| 29 | Apply missing metadata.name | Error path | milestone-2-apply.feature | TestPolicyApply_MissingName |
| 30 | Apply invalid duration | Error path | milestone-2-apply.feature | TestPolicyApply_InvalidDuration |
| 31 | Apply zero selectors on entry | Error path | milestone-2-apply.feature | TestPolicyApply_ZeroSelectors |
| 32 | Apply multiple selectors on entry | Error path | milestone-2-apply.feature | TestPolicyApply_MultipleSelectorsOnEntry |
| 33 | Apply empty access list | Error path | milestone-2-apply.feature | TestPolicyApply_EmptyAccess |
| 34 | Apply multiple validation errors | Error path | milestone-2-apply.feature | TestPolicyApply_MultipleErrors |
| 35 | Apply file not found | Error path | milestone-2-apply.feature | TestPolicyApply_FileNotFound |
| 36 | Apply invalid YAML | Error path | milestone-2-apply.feature | TestPolicyApply_InvalidYAML |
| 37 | Apply no file argument | Error path | milestone-2-apply.feature | TestPolicyApply_NoFileArg |

### US-PM-04: Delete Policy

| # | Scenario | Type | Feature File | Go Test Function |
|---|---|---|---|---|
| 38 | Delete with --yes | Happy path | milestone-3-delete-init.feature | TestPolicyDelete_WithYes |
| 39 | Delete interactive confirm | Happy path | milestone-3-delete-init.feature | TestPolicyDelete_InteractiveYes |
| 40 | Delete cancelled by user | Edge case | milestone-3-delete-init.feature | TestPolicyDelete_Cancelled |
| 41 | Delete non-existent | Error path | milestone-3-delete-init.feature | TestPolicyDelete_NotFound |
| 42 | Delete JSON output | Happy path | milestone-3-delete-init.feature | TestPolicyDelete_JSON |

### US-PM-05: Init Policy Scaffold

| # | Scenario | Type | Feature File | Go Test Function |
|---|---|---|---|---|
| 43 | Scaffold new policy | Happy path | milestone-3-delete-init.feature | TestPolicyInit_NewFile |
| 44 | Scaffold warns existing file | Edge case | milestone-3-delete-init.feature | TestPolicyInit_ExistingFile |
| 45 | Scaffold produces valid YAML | Happy path | milestone-3-delete-init.feature | TestPolicyInit_ValidYAML |
| 46 | Init works offline | Edge case | milestone-3-delete-init.feature | TestPolicyInit_NoNetwork |

### US-PM-06: Who-Uses

| # | Scenario | Type | Feature File | Go Test Function |
|---|---|---|---|---|
| 47 | Who-uses by name | Happy path | milestone-4-phase2.feature | TestWhoUses_ByName |
| 48 | Who-uses by ID | Happy path | milestone-4-phase2.feature | TestWhoUses_ByID |
| 49 | Who-uses by listenPath | Happy path | milestone-4-phase2.feature | TestWhoUses_ByPath |
| 50 | Who-uses no references | Edge case | milestone-4-phase2.feature | TestWhoUses_NoReferences |
| 51 | Who-uses non-existent API | Error path | milestone-4-phase2.feature | TestWhoUses_APINotFound |
| 52 | Who-uses JSON output | Happy path | milestone-4-phase2.feature | TestWhoUses_JSON |

### US-PM-07: Bind/Unbind

| # | Scenario | Type | Feature File | Go Test Function |
|---|---|---|---|---|
| 53 | Bind API to policy | Happy path | milestone-4-phase2.feature | TestPolicyBind_Success |
| 54 | Bind default version | Happy path | milestone-4-phase2.feature | TestPolicyBind_DefaultVersion |
| 55 | Bind already bound | Error path | milestone-4-phase2.feature | TestPolicyBind_AlreadyBound |
| 56 | Bind policy not found | Error path | milestone-4-phase2.feature | TestPolicyBind_PolicyNotFound |
| 57 | Bind API not found | Error path | milestone-4-phase2.feature | TestPolicyBind_APINotFound |
| 58 | Unbind API from policy | Happy path | milestone-4-phase2.feature | TestPolicyUnbind_Success |
| 59 | Unbind API not in policy | Error path | milestone-4-phase2.feature | TestPolicyUnbind_NotInPolicy |
| 60 | Unbind policy not found | Error path | milestone-4-phase2.feature | TestPolicyUnbind_PolicyNotFound |

---

## Unit Test Inventory (Supporting Inner-Loop TDD)

These are not Gherkin scenarios but table-driven unit tests that the software-crafter writes as part of the inner loop.

### internal/policy/duration_test.go

| Test | Cases |
|---|---|
| TestParseDuration | "30d"->2592000, "24h"->86400, "1m"->60, "60s"->60, "60"->60, "0"->0 |
| TestParseDuration_Errors | "abc", "-1", "1.5h", "1h30m", "", "30w" |
| TestFormatDuration | 2592000->"30d", 86400->"1d", 3600->"1h", 60->"1m", 45->"45s", 0->0 |

### internal/policy/validate_test.go

| Test | Cases |
|---|---|
| TestValidatePolicy_Valid | Complete valid PolicyFile |
| TestValidatePolicy_MissingID | metadata.id empty |
| TestValidatePolicy_MissingName | metadata.name empty |
| TestValidatePolicy_InvalidDuration | per: "abc" |
| TestValidatePolicy_NoAccess | empty access list |
| TestValidatePolicy_ZeroSelectors | access entry with no selector fields |
| TestValidatePolicy_MultipleSelectors | access entry with name + id both set |
| TestValidatePolicy_MultipleErrors | collects all errors in one pass |

### internal/policy/selector_test.go

| Test | Cases |
|---|---|
| TestResolveByName_Exact | "users-api" -> a1b2c3d4e5f6 |
| TestResolveByName_NotFound | "inventori-api" -> error with suggestions |
| TestResolveByName_Ambiguous | "api-service" matches 2 -> error with candidates |
| TestResolveByListenPath_Exact | "/orders/" -> g7h8i9j0k1l2 |
| TestResolveByListenPath_Ambiguous | "/shared/" matches 2 -> error |
| TestResolveByID_Found | "a1b2c3d4e5f6" -> direct |
| TestResolveByID_NotFound | "nonexistent" -> error |
| TestResolveByTags_MultiMatch | [public, v1] -> 2 APIs |
| TestResolveByTags_NoMatch | [internal, legacy] -> error |
| TestResolveAll_MixedSelectors | name + listenPath + tags in one policy |
| TestFuzzySuggestions | "inventori-api" suggests "inventory-api" |

### internal/policy/convert_test.go

| Test | Cases |
|---|---|
| TestCLIToWire | Full PolicyFile -> DashboardPolicy field mapping |
| TestWireToCLI | Full DashboardPolicy -> PolicyFile field mapping |
| TestRoundTrip | CLI->wire->CLI produces equivalent content |
| TestAccessConversion | access entries -> access_rights map and back |

### internal/client/policy_test.go

| Test | Cases |
|---|---|
| TestListPolicies | GET /api/portal/policies?p=1 + response parsing |
| TestListPolicies_Empty | Empty list response |
| TestGetPolicy | GET /api/portal/policies/{id} + response parsing |
| TestGetPolicy_NotFound | 404 -> ErrorResponse |
| TestCreatePolicy | POST /api/portal/policies + request body verification |
| TestUpdatePolicy | PUT /api/portal/policies/{id} + request body verification |
| TestDeletePolicy | DELETE /api/portal/policies/{id} |
| TestDeletePolicy_NotFound | 404 -> ErrorResponse |

### pkg/types/policy_test.go

| Test | Cases |
|---|---|
| TestPolicyFile_YAMLRoundTrip | Marshal -> unmarshal -> equivalent |
| TestPolicyFile_JSONRoundTrip | Marshal -> unmarshal -> equivalent |
| TestDashboardPolicy_JSONRoundTrip | Marshal -> unmarshal -> equivalent |
| TestDuration_UnmarshalYAML | String "30d" and integer 60 both accepted |

---

## Implementation Sequence (One-at-a-Time)

Enabled test order for the software-crafter:

```
 1. TestPolicyList_Empty                    (walking skeleton 1a)
 2. TestPolicyList_WithPolicies             (walking skeleton 1b)
 3. TestPolicyApply_Create_NameSelector     (walking skeleton 2a)
 4. TestPolicyApply_Update_Idempotent       (walking skeleton 2b)
 --- walking skeleton complete, all layers proven ---
 5. TestPolicyList_JSONOutput
 6. TestPolicyList_Pagination_EmptyPage
 7. TestPolicyList_APICount
 8. TestPolicyList_Tags
 9. TestPolicyGet_Human
10. TestPolicyGet_JSON
11. TestPolicyGet_NotFound
12. TestPolicyApply_ListenPathSelector
13. TestPolicyApply_IDSelector
14. TestPolicyApply_TagsSelector
15. TestPolicyApply_DurationConversion
16. TestPolicyApply_NameNotFound
17. TestPolicyApply_NameAmbiguous
18. TestPolicyApply_MissingID
19. TestPolicyApply_InvalidDuration
20. TestPolicyApply_EmptyAccess
21. TestPolicyApply_FileNotFound
22. TestPolicyDelete_WithYes
23. TestPolicyDelete_NotFound
24. TestPolicyInit_NewFile
... (remaining scenarios)
```

Each test is enabled one at a time by removing the `t.Skip("pending")` marker.
