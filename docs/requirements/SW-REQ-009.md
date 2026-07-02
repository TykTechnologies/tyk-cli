# SW-REQ-009: GenerateListenPath falls back to '/api/' when the title sanitises to an empty slug, ensuring listen_path_valid even for degenerate input

<!-- reqproof:req SW-REQ-009 -->

## Intent
GenerateListenPath falls back to '/api/' when the title sanitises to an empty slug, ensuring listen_path_valid even for degenerate input

## Context
- Level: Software
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: SYS-REQ-010

## Formalization
```
when title_sanitises_to_empty the api shall immediately satisfy (result_uses_api_fallback & listen_path_valid)
```

## Source
See `specs/software/requirements/SW-REQ-009.req.yaml` for the machine-readable definition.
