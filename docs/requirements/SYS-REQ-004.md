# SYS-REQ-004: api import-oas accepts a local file (--file) or remote URL (--url) containing a plain OAS document and creates a new API, always generating a new ID

<!-- reqproof:req SYS-REQ-004 -->

## Intent
api import-oas accepts a local file (--file) or remote URL (--url) containing a plain OAS document and creates a new API, always generating a new ID

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when import_oas_invoked the api shall immediately satisfy import_takes_create_path
```

## Source
See `specs/system/requirements/SYS-REQ-004.req.yaml` for the machine-readable definition.
