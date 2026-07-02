# SYS-REQ-016: api create, api import-oas, and api apply exit 4 when the Dashboard returns a conflict (HTTP 409)

<!-- reqproof:req SYS-REQ-016 -->

## Intent
api create, api import-oas, and api apply exit 4 when the Dashboard returns a conflict (HTTP 409)

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when api_conflict_response the api shall immediately satisfy api_exit_conflict
```

## Source
See `specs/system/requirements/SYS-REQ-016.req.yaml` for the machine-readable definition.
