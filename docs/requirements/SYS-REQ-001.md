# SYS-REQ-001: api list returns paginated OAS APIs from the Dashboard; page defaults to 1, 10 per page; supports --json and --interactive flags

<!-- reqproof:req SYS-REQ-001 -->

## Intent
api list returns paginated OAS APIs from the Dashboard; page defaults to 1, 10 per page; supports --json and --interactive flags

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when page_flag_omitted the api shall immediately satisfy requested_page_equals_one
```

## Source
See `specs/system/requirements/SYS-REQ-001.req.yaml` for the machine-readable definition.
