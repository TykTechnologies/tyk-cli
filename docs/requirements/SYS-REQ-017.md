# SYS-REQ-017: API subcommands map authentication failures (HTTP 401) to exit 5, so callers can detect 'rotate the auth token' situations programmatically

<!-- reqproof:req SYS-REQ-017 -->

## Intent
API subcommands map authentication failures (HTTP 401) to exit 5, so callers can detect 'rotate the auth token' situations programmatically

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: STK-REQ-001

## Formalization
```
when api_auth_failed_response the api shall immediately satisfy api_exit_auth_failed
```

## Source
See `specs/system/requirements/SYS-REQ-017.req.yaml` for the machine-readable definition.
