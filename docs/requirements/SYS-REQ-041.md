# SYS-REQ-041: 'Configuration is resolved in priority order: CLI flags override environment variables, which override named environments in ~/.config/tyk/cli.toml'

<!-- reqproof:req SYS-REQ-041 -->

## Intent
'Configuration is resolved in priority order: CLI flags override environment variables, which override named environments in ~/.config/tyk/cli.toml'

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when flag_value_provided the config shall immediately satisfy effective_value_from_flag
```

## Source
See `specs/system/requirements/SYS-REQ-041.req.yaml` for the machine-readable definition.
