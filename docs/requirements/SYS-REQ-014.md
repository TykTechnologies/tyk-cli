# SYS-REQ-014: All api subcommands exit 2 when required flags are missing or arguments are invalid (before any network call)

<!-- reqproof:req SYS-REQ-014 -->

## Intent
All api subcommands exit 2 when required flags are missing or arguments are invalid (before any network call)

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when api_invalid_args the api shall immediately satisfy api_exit_bad_args
```

## Source
See `specs/system/requirements/SYS-REQ-014.req.yaml` for the machine-readable definition.
