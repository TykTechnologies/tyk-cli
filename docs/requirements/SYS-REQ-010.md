# SYS-REQ-010: GenerateListenPath converts an API title to a URL-safe slug path (lowercase, non-alphanumeric replaced with hyphens, wrapped in leading/trailing slashes)

<!-- reqproof:req SYS-REQ-010 -->

## Intent
GenerateListenPath converts an API title to a URL-safe slug path (lowercase, non-alphanumeric replaced with hyphens, wrapped in leading/trailing slashes)

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when api_title_provided the api shall immediately satisfy listen_path_valid
```

## Source
See `specs/system/requirements/SYS-REQ-010.req.yaml` for the machine-readable definition.
