# SYS-REQ-042: tyk init runs an interactive wizard that writes one or more named environments to ~/.config/tyk/cli.toml and sets the default environment

<!-- reqproof:req SYS-REQ-042 -->

## Intent
tyk init runs an interactive wizard that writes one or more named environments to ~/.config/tyk/cli.toml and sets the default environment

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when init_wizard_completed the config shall immediately satisfy (toml_file_written & default_environment_set)
```

## Source
See `specs/system/requirements/SYS-REQ-042.req.yaml` for the machine-readable definition.
