# SYS-REQ-036: policy apply exits 4 on a Dashboard conflict (HTTP 409)

<!-- reqproof:req SYS-REQ-036 -->

## Intent
policy apply exits 4 on a Dashboard conflict (HTTP 409)

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when policy_conflict_response the policy shall immediately satisfy policy_exit_conflict
```

## Source
See `specs/system/requirements/SYS-REQ-036.req.yaml` for the machine-readable definition.
