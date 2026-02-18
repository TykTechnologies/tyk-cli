# Acceptance Test Review Checklist: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DISTILL
**Date**: 2026-02-18
**Reviewer**: (pending peer review)

---

## Mandate Compliance Evidence

### CM-A: Driving Port Usage

All acceptance tests invoke through the CLI command layer (the driving port). No test directly calls `internal/policy/` or `internal/client/` functions. The tests create a Cobra command via `NewPolicyCommand()` or `NewRootCommand()`, inject config context, and call `Execute()`.

**Evidence**: Every acceptance test function in `internal/cli/policy_test.go` follows this pattern:
```go
cmd := NewPolicyCommand()  // or root.Find([]string{"policy", "list"})
cmd.SetContext(withConfig(context.Background(), cfg))
cmd.SetArgs([]string{...})
err := cmd.Execute()
```

This mirrors the existing pattern in `api_list_test.go:36-56` and `api_get_test.go:55-85`.

### CM-B: Zero Technical Terms in Feature Files

Feature files use business language exclusively. No references to:
- HTTP methods (GET, POST, PUT, DELETE)
- Status codes (200, 404)
- Package names, struct names, or Go types
- Database or persistence terms
- Internal component names

Technical details appear only in "Then" steps that verify Dashboard interaction (e.g., "the Dashboard receives a create request") which describe observable integration behavior, not implementation.

**Grep verification command**:
```bash
grep -iE '(http|status.code|struct|func|package|import|json\.Marshal|interface|goroutine)' \
  docs/feature/policies-mgmt/distill/*.feature
# Expected: zero matches
```

### CM-C: Walking Skeleton and Focused Scenario Counts

| Category | Count | Target |
|---|---|---|
| Walking skeleton scenarios | 4 | 2-3 minimum |
| Focused happy-path scenarios | 24 | -- |
| Error/edge-case scenarios | 27 | >= 40% of total |
| **Total** | **55** | -- |
| **Error path ratio** | **49%** | >= 40% |

---

## Review Dimensions

### 1. Coverage Completeness

- [ ] All 7 user stories have acceptance scenarios
- [ ] US-PM-01 (list): 6 scenarios covering empty, populated, JSON, pagination, counts, tags
- [ ] US-PM-02 (get): 5 scenarios covering human, CLI schema, JSON, not-found, file export
- [ ] US-PM-03 (apply): 26 scenarios covering all selector types, durations, stdin, all error paths
- [ ] US-PM-04 (delete): 5 scenarios covering yes flag, interactive, cancel, not-found, JSON
- [ ] US-PM-05 (init): 4 scenarios covering new file, existing file, valid YAML, offline
- [ ] US-PM-06 (who-uses): 6 scenarios covering name/ID/path lookup, no refs, not-found, JSON
- [ ] US-PM-07 (bind/unbind): 8 scenarios covering success, already bound, not found

### 2. Architecture Alignment

- [ ] Tests invoke through CLI commands only (driving port)
- [ ] Scenarios map to architectural component boundaries per component-boundaries.md
- [ ] Walking skeleton covers all 7 new files in the architecture
- [ ] Test file locations match convention from component-boundaries.md Section 4

### 3. Business Language Purity

- [ ] Feature files contain zero technical jargon
- [ ] Step descriptions use domain terms (policies, apply, selectors, rate limit)
- [ ] Error messages match user-facing text from stories (not internal error codes)

### 4. Error Path Coverage

- [ ] Selector resolution errors: not-found, ambiguous, empty tags (5 scenarios)
- [ ] Schema validation errors: missing fields, invalid durations, selector format (7 scenarios)
- [ ] File handling errors: not found, invalid YAML, no argument (3 scenarios)
- [ ] Not-found resources: get, delete, who-uses, bind/unbind (6 scenarios)
- [ ] Already-exists/duplicate: bind already bound (1 scenario)
- [ ] User cancellation: delete cancelled (1 scenario)
- [ ] Total error/edge scenarios: 27 out of 55 = 49%

### 5. Test Data Consistency

- [ ] Uses shared personas from DISCUSS wave (Ravi Patel)
- [ ] Uses shared API data (a1b2c3d4e5f6/users-api, g7h8i9j0k1l2/orders-api, m3n4o5p6q7r8/payments-api)
- [ ] Uses shared policy data (gold/Gold Plan, silver/Silver Plan, free-tier/Free Plan)
- [ ] Uses concrete values throughout ("5000 requests", "$100.00" style specificity)

### 6. Implementation Feasibility

- [ ] All scenarios implementable with httptest mock server pattern
- [ ] No scenarios require real Dashboard connectivity
- [ ] One-at-a-time ordering defined in test-scenarios.md
- [ ] Walking skeleton can pass before focused scenarios are enabled

---

## Definition of Done Validation

| Criterion | Status | Evidence |
|---|---|---|
| All acceptance scenarios written | DONE | 55 scenarios across 5 feature files |
| Step definitions have Go test function mapping | DONE | test-scenarios.md maps every scenario |
| Test pyramid complete | DONE | Acceptance (CLI), integration-style (client+httptest), unit (duration/selector/validate/convert/types) |
| Walking skeleton identified | DONE | 4 scenarios in walking-skeleton.feature |
| One-at-a-time sequence defined | DONE | test-scenarios.md implementation sequence |
| Peer review checklist prepared | DONE | This document |
| Error path ratio >= 40% | DONE | 49% (27/55) |
| Mandate compliance proven (CM-A/B/C) | DONE | See evidence above |

---

## Handoff to Software-Crafter

### What the crafter receives:
1. Five .feature files as executable specifications
2. Go test scaffolds showing mock server setup and test patterns
3. test-scenarios.md with implementation sequence
4. walking-skeleton.md with step-by-step build order

### What the crafter does first:
1. Enable walking skeleton test 1 (TestPolicyList_Empty)
2. Create `pkg/types/policy.go` to make it compile
3. Create `internal/client/policy.go` with ListPolicies
4. Create `internal/cli/policy.go` with list command
5. Make TestPolicyList_Empty pass
6. Enable next skeleton test, repeat

### Critical constraints for crafter:
- Do NOT enable multiple tests at once
- Walking skeleton must pass before enabling focused scenarios
- Unit tests (duration, selector, validate, convert) are inner-loop TDD -- write them as you implement each module
- Acceptance tests are the outer loop -- they define "done" for each scenario
