# SYS-REQ-015: api get and api delete exit 3 when the target API ID is not found (HTTP 404)

<!-- reqproof:req SYS-REQ-015 -->

## Intent
api get and api delete exit 3 when the target API ID is not found (HTTP 404)

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when api_target_not_found the api shall immediately satisfy api_exit_not_found
```

## Source
See `specs/system/requirements/SYS-REQ-015.req.yaml` for the machine-readable definition.
