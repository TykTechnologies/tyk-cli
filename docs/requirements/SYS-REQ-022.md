# SYS-REQ-022: api commands load OAS documents from a local file path (YAML or JSON) or by fetching a URL, returning a normalized map structure

<!-- reqproof:req SYS-REQ-022 -->

## Intent
api commands load OAS documents from a local file path (YAML or JSON) or by fetching a URL, returning a normalized map structure

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (oas_file_provided | oas_url_provided) the api shall immediately satisfy parsed_oas_returned
```

## Source
See `specs/system/requirements/SYS-REQ-022.req.yaml` for the machine-readable definition.
