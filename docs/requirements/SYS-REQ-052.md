# SYS-REQ-052: config manager supports CRUD operations for named environments and default environment selection, persisting changes to the TOML config file

<!-- reqproof:req SYS-REQ-052 -->

## Intent
config manager supports CRUD operations for named environments and default environment selection, persisting changes to the TOML config file

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when crud_op_invoked the config shall immediately satisfy toml_persisted
```

## Source
See `specs/system/requirements/SYS-REQ-052.req.yaml` for the machine-readable definition.
