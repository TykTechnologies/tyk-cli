# SYS-REQ-049: All config subcommands exit 0 on success

<!-- reqproof:req SYS-REQ-049 -->

## Intent
All config subcommands exit 0 on success

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when config_operation_succeeded the config shall immediately satisfy config_exit_zero
```

## Source
See `specs/system/requirements/SYS-REQ-049.req.yaml` for the machine-readable definition.
