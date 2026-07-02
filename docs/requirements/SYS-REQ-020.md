# SYS-REQ-020: API subcommands map authorization failures (HTTP 403) to exit 6, distinguishing 'token lacks permission' from authentication failures

<!-- reqproof:req SYS-REQ-020 -->

## Intent
API subcommands map authorization failures (HTTP 403) to exit 6, distinguishing 'token lacks permission' from authentication failures

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
when api_forbidden_response the api shall immediately satisfy api_exit_forbidden
```

## Source
See `specs/system/requirements/SYS-REQ-020.req.yaml` for the machine-readable definition.
