# SYS-REQ-037: POL subcommands map authentication failures (HTTP 401) to exit 5, so callers can detect 'rotate the auth token' situations programmatically

<!-- reqproof:req SYS-REQ-037 -->

## Intent
POL subcommands map authentication failures (HTTP 401) to exit 5, so callers can detect 'rotate the auth token' situations programmatically

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
when policy_auth_failed_response the policy shall immediately satisfy policy_exit_auth_failed
```

## Source
See `specs/system/requirements/SYS-REQ-037.req.yaml` for the machine-readable definition.
