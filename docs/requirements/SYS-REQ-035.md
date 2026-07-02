# SYS-REQ-035: policy get and policy delete exit 3 when the target policy ID is not found (HTTP 404)

<!-- reqproof:req SYS-REQ-035 -->

## Intent
policy get and policy delete exit 3 when the target policy ID is not found (HTTP 404)

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when policy_target_not_found the policy shall immediately satisfy policy_exit_not_found
```

## Source
See `specs/system/requirements/SYS-REQ-035.req.yaml` for the machine-readable definition.
