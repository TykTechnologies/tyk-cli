# SYS-REQ-018: API subcommands map rate-limit responses (HTTP 429) to exit 7, enabling CI scripts to detect throttling and back off

<!-- reqproof:req SYS-REQ-018 -->

## Intent
API subcommands map rate-limit responses (HTTP 429) to exit 7, enabling CI scripts to detect throttling and back off

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
when api_rate_limited_response the api shall immediately satisfy api_exit_rate_limited
```

## Source
See `specs/system/requirements/SYS-REQ-018.req.yaml` for the machine-readable definition.
