# SYS-REQ-050: All config subcommands exit 2 on invalid arguments or missing required fields

<!-- reqproof:req SYS-REQ-050 -->

## Intent
All config subcommands exit 2 on invalid arguments or missing required fields

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when config_invalid_args the config shall immediately satisfy config_exit_bad_args
```

## Source
See `specs/system/requirements/SYS-REQ-050.req.yaml` for the machine-readable definition.
