# SYS-REQ-038: POL subcommands map rate-limit responses (HTTP 429) to exit 7, enabling CI scripts to detect throttling and back off

<!-- reqproof:req SYS-REQ-038 -->

## Intent
POL subcommands map rate-limit responses (HTTP 429) to exit 7, enabling CI scripts to detect throttling and back off

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: STK-REQ-001

## Formalization
```
when policy_rate_limited_response the policy shall immediately satisfy policy_exit_rate_limited
```

## Source
See `specs/system/requirements/SYS-REQ-038.req.yaml` for the machine-readable definition.
