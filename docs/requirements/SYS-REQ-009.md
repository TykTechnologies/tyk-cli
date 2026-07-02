# SYS-REQ-009: AddTykExtensions injects minimal x-tyk-api-gateway block into a plain OAS document; it is a no-op if extensions are already present; it fails if info.title or servers[0].url is missing

<!-- reqproof:req SYS-REQ-009 -->

## Intent
AddTykExtensions injects minimal x-tyk-api-gateway block into a plain OAS document; it is a no-op if extensions are already present; it fails if info.title or servers[0].url is missing

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when oas_lacks_extensions the api shall immediately satisfy extensions_added
```

## Source
See `specs/system/requirements/SYS-REQ-009.req.yaml` for the machine-readable definition.
