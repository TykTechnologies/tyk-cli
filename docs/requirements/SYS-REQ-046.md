# SYS-REQ-046: tyk config remove deletes a named environment from cli.toml; refuses to remove an unknown environment, and refuses to remove the sole remaining environment

<!-- reqproof:req SYS-REQ-046 -->

## Intent
tyk config remove deletes a named environment from cli.toml; refuses to remove an unknown environment, and refuses to remove the sole remaining environment

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: STK-REQ-003

## Formalization
```
when (remove_targets_only_environment | remove_targets_unknown_environment) the config shall immediately satisfy (remove_last_environment_rejected | remove_unknown_environment_rejected)
```

## Source
See `specs/system/requirements/SYS-REQ-046.req.yaml` for the machine-readable definition.
