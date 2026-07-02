# SYS-REQ-006: api update-oas updates an existing API's OAS document while preserving its existing x-tyk-api-gateway extensions; the API must already exist or the command exits with code 3

<!-- reqproof:req SYS-REQ-006 -->

## Intent
api update-oas updates an existing API's OAS document while preserving its existing x-tyk-api-gateway extensions; the API must already exist or the command exits with code 3

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (update_invoked & update_target_does_not_exist) the api shall immediately satisfy api_exit_not_found
```

## Source
See `specs/system/requirements/SYS-REQ-006.req.yaml` for the machine-readable definition.
