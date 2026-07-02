# SYS-REQ-011: OAS document utilities detect x-tyk-api-gateway presence, extract API ID and upstream URL, and check document structure before transformation

<!-- reqproof:req SYS-REQ-011 -->

## Intent
OAS document utilities detect x-tyk-api-gateway presence, extract API ID and upstream URL, and check document structure before transformation

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when oas_has_tyk_extensions the api shall immediately satisfy api_id_extractable
```

## Source
See `specs/system/requirements/SYS-REQ-011.req.yaml` for the machine-readable definition.
