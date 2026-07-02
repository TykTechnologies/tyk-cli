# SYS-REQ-005: 'api apply performs an idempotent upsert: if x-tyk-api-gateway.info.id is present it updates the existing API (or creates with that ID if not found); if absent it creates a new API'

<!-- reqproof:req SYS-REQ-005 -->

## Intent
'api apply performs an idempotent upsert: if x-tyk-api-gateway.info.id is present it updates the existing API (or creates with that ID if not found); if absent it creates a new API'

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (apply_invoked & apply_target_exists) the api shall immediately satisfy apply_takes_update_path
```

## Source
See `specs/system/requirements/SYS-REQ-005.req.yaml` for the machine-readable definition.
