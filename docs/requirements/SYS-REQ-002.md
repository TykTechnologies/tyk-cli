# SYS-REQ-002: api get retrieves a single API by ID; --oas-only strips x-tyk-api-gateway extensions from output; --version-name selects a specific version

<!-- reqproof:req SYS-REQ-002 -->

## Intent
api get retrieves a single API by ID; --oas-only strips x-tyk-api-gateway extensions from output; --version-name selects a specific version

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when oas_only_flag_set the api shall immediately satisfy tyk_extensions_stripped
```

## Source
See `specs/system/requirements/SYS-REQ-002.req.yaml` for the machine-readable definition.
