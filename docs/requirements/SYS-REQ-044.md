# SYS-REQ-044: tyk config list renders every configured environment and marks the active one; emits a friendly empty-state message when no environments are configured

<!-- reqproof:req SYS-REQ-044 -->

## Intent
tyk config list renders every configured environment and marks the active one; emits a friendly empty-state message when no environments are configured

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
when list_invoked_with_environments_present the config shall immediately satisfy all_environments_rendered_with_active_marked
```

## Source
See `specs/system/requirements/SYS-REQ-044.req.yaml` for the machine-readable definition.
