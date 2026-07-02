# SYS-REQ-012: 'API runners (import-oas, update-oas) validate the OAS document structure locally before any Dashboard round-trip: openapi field is non-empty 3.x, info section exists with non-empty title and version. Missing fields produce exit 2 (ExitBadArgs) without contacting the Dashboard.'

<!-- reqproof:req SYS-REQ-012 -->

## Intent
'API runners (import-oas, update-oas) validate the OAS document structure locally before any Dashboard round-trip: openapi field is non-empty 3.x, info section exists with non-empty title and version. Missing fields produce exit 2 (ExitBadArgs) without contacting the Dashboard.'

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: STK-REQ-001

## Formalization
```
when oas_missing_required_fields the api shall immediately satisfy (dashboard_round_trip_aborted & api_exit_bad_args)
```

## Source
See `specs/system/requirements/SYS-REQ-012.req.yaml` for the machine-readable definition.
