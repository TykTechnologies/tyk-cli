# SYS-REQ-040: POL subcommands map authorization failures (HTTP 403) to exit 6, distinguishing 'token lacks permission' from authentication failures

<!-- reqproof:req SYS-REQ-040 -->

## Intent
POL subcommands map authorization failures (HTTP 403) to exit 6, distinguishing 'token lacks permission' from authentication failures

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
when policy_forbidden_response the policy shall immediately satisfy policy_exit_forbidden
```

## Source
See `specs/system/requirements/SYS-REQ-040.req.yaml` for the machine-readable definition.
