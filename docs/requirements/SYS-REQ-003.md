# SYS-REQ-003: api create builds a minimal OAS document with x-tyk-api-gateway extensions from --name and --upstream-url, auto-generating listen path when --listen-path is omitted

<!-- reqproof:req SYS-REQ-003 -->

## Intent
api create builds a minimal OAS document with x-tyk-api-gateway extensions from --name and --upstream-url, auto-generating listen path when --listen-path is omitted

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (listen_path_flag_omitted & api_title_provided) the api shall immediately satisfy listen_path_auto_generated
```

## Source
See `specs/system/requirements/SYS-REQ-003.req.yaml` for the machine-readable definition.
