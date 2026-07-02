# SYS-REQ-043: config use switches the active named environment; the switch persists to ~/.config/tyk/cli.toml; config current shows the active environment name and its values

<!-- reqproof:req SYS-REQ-043 -->

## Intent
config use switches the active named environment; the switch persists to ~/.config/tyk/cli.toml; config current shows the active environment name and its values

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when use_invoked_with_valid_env the config shall immediately satisfy (default_env_switched & toml_persisted_with_new_default)
```

## Source
See `specs/system/requirements/SYS-REQ-043.req.yaml` for the machine-readable definition.
