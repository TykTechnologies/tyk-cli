# SYS-REQ-019: API subcommands map server errors (HTTP 5xx) to exit 8, distinguishing Dashboard-side failures from local errors

<!-- reqproof:req SYS-REQ-019 -->

## Intent
API subcommands map server errors (HTTP 5xx) to exit 8, distinguishing Dashboard-side failures from local errors

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
when api_server_error_response the api shall immediately satisfy api_exit_server_error
```

## Source
See `specs/system/requirements/SYS-REQ-019.req.yaml` for the machine-readable definition.
