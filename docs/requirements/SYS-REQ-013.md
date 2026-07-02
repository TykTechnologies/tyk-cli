# SYS-REQ-013: All api subcommands exit 0 on success

<!-- reqproof:req SYS-REQ-013 -->

## Intent
All api subcommands exit 0 on success

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when api_operation_succeeded the api shall immediately satisfy api_exit_zero
```

## Source
See `specs/system/requirements/SYS-REQ-013.req.yaml` for the machine-readable definition.
