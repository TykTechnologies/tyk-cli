# Walking Skeleton Strategy: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DISTILL
**Date**: 2026-02-18

---

## Walking Skeleton Definition

The walking skeleton consists of 4 scenarios that validate the complete vertical slice through every architectural layer:

### Skeleton 1: List policies (read path)
- **Scenario**: "Ravi lists policies and sees an empty inventory"
- **Scenario**: "Ravi lists policies and sees existing policies"
- **Layers exercised**: CLI command -> client.ListPolicies -> httptest mock -> response parsing -> table output

### Skeleton 2: Apply a new policy (write path)
- **Scenario**: "Ravi applies a new policy with name-based selectors"
- **Scenario**: "Ravi updates an existing policy idempotently"
- **Layers exercised**: CLI command -> filehandler.LoadFile -> validate.ValidatePolicy -> selector.ResolveAll -> duration.ParseDuration -> convert.CLIToWire -> client.CreatePolicy/UpdatePolicy -> httptest mock -> success output

## Why These 4 Scenarios

| Concern | List scenarios | Apply scenarios |
|---|---|---|
| Cobra command registration | Yes | Yes |
| Config context extraction | Yes | Yes |
| Client construction | Yes | Yes |
| HTTP request formation | GET /api/portal/policies | POST + PUT /api/portal/policies |
| Response deserialization | Yes (DashboardPolicyListResponse) | Yes (DashboardPolicy) |
| Output formatting (table) | Yes | - |
| File loading (YAML) | - | Yes |
| Schema validation | - | Yes |
| Selector resolution | - | Yes (name -> API ID) |
| Duration parsing | - | Yes (1m -> 60, 30d -> 2592000) |
| CLI-to-wire conversion | - | Yes |
| Upsert logic (create vs update) | - | Yes (both paths) |
| ExitError codes | - | Covered in focused tests |

Together, these 4 scenarios exercise every new file in the architecture:
- `internal/cli/policy.go`
- `internal/client/policy.go`
- `internal/policy/selector.go`
- `internal/policy/duration.go`
- `internal/policy/validate.go`
- `internal/policy/convert.go`
- `pkg/types/policy.go`

## Stakeholder Demo Capability

Each walking skeleton is demo-able to stakeholders:

1. **"Can I see what policies exist?"** -- Run `tyk policy list`, see a table or "No policies found."
2. **"Can I push a policy from a YAML file?"** -- Run `tyk policy apply -f policies/platinum.yaml`, see resolution log and "Policy applied successfully!"
3. **"Is it safe to run twice?"** -- Run the same apply again, see "Status: updated" instead of creating a duplicate.

## Implementation Order

```
Step 1: pkg/types/policy.go            -- types that all layers share
Step 2: internal/policy/duration.go     -- pure function, no deps
Step 3: internal/policy/validate.go     -- depends on types only
Step 4: internal/policy/selector.go     -- depends on types only (mock API list)
Step 5: internal/policy/convert.go      -- depends on types + duration
Step 6: internal/client/policy.go       -- depends on types + existing client
Step 7: internal/cli/policy.go          -- wires everything, walking skeleton passes
Step 8: internal/cli/root.go            -- 1-line registration
```

At Step 7, the walking skeleton test (4 scenarios) should pass. All other scenarios remain @pending.

## Go Test File Mapping

| Feature File | Go Test File | Test Layer |
|---|---|---|
| walking-skeleton.feature | `internal/cli/policy_test.go` | Acceptance (full command + httptest) |
| milestone-1-list-get.feature | `internal/cli/policy_test.go` | Acceptance (full command + httptest) |
| milestone-2-apply.feature | `internal/cli/policy_test.go` | Acceptance (full command + httptest) |
| milestone-3-delete-init.feature | `internal/cli/policy_test.go` | Acceptance (full command + httptest) |
| milestone-4-phase2.feature | `internal/cli/policy_test.go` | Acceptance (Phase 2) |
| (supporting unit tests) | `internal/policy/duration_test.go` | Unit |
| (supporting unit tests) | `internal/policy/selector_test.go` | Unit |
| (supporting unit tests) | `internal/policy/validate_test.go` | Unit |
| (supporting unit tests) | `internal/policy/convert_test.go` | Unit |
| (supporting unit tests) | `internal/client/policy_test.go` | Integration-style unit |
| (supporting unit tests) | `pkg/types/policy_test.go` | Unit |
