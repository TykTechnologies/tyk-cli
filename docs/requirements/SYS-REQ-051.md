# SYS-REQ-051: config manager loads environments from a TOML file in the XDG base directory, falls back to environment variables, and exposes an effective config resolving precedence

<!-- reqproof:req SYS-REQ-051 -->

## Intent
config manager loads environments from a TOML file in the XDG base directory, falls back to environment variables, and exposes an effective config resolving precedence

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when toml_file_exists the config shall immediately satisfy config_loaded_from_toml
```

## Source
See `specs/system/requirements/SYS-REQ-051.req.yaml` for the machine-readable definition.
