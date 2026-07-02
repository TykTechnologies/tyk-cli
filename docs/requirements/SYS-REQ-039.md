# SYS-REQ-039: POL subcommands map server errors (HTTP 5xx) to exit 8, distinguishing Dashboard-side failures from local errors

<!-- reqproof:req SYS-REQ-039 -->

## Intent
POL subcommands map server errors (HTTP 5xx) to exit 8, distinguishing Dashboard-side failures from local errors

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
when policy_server_error_response the policy shall immediately satisfy policy_exit_server_error
```

## Source
See `specs/system/requirements/SYS-REQ-039.req.yaml` for the machine-readable definition.
