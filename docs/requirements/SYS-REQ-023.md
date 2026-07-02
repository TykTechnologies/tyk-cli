# SYS-REQ-023: tyk api versions list/create/switch-default are reserved placeholder subcommands for a future Phase 3 implementation; current invocation emits a fixed placeholder message and returns without contacting the Dashboard. This requirement documents the placeholder contract so the surface is auditable until full implementation lands.

<!-- reqproof:req SYS-REQ-023 -->

## Intent
tyk api versions list/create/switch-default are reserved placeholder subcommands for a future Phase 3 implementation; current invocation emits a fixed placeholder message and returns without contacting the Dashboard. This requirement documents the placeholder contract so the surface is auditable until full implementation lands.

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: STK-REQ-001

## Formalization
```
when versions_subcommand_invoked the api shall immediately satisfy (placeholder_message_emitted & no_dashboard_contact_made)
```

## Source
See `specs/system/requirements/SYS-REQ-023.req.yaml` for the machine-readable definition.
