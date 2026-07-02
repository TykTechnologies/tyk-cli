# SYS-REQ-029: policy YAML is validated for required fields (friendly ID, access rights) and structural constraints before being submitted to the Dashboard

<!-- reqproof:req SYS-REQ-029 -->

## Intent
policy YAML is validated for required fields (friendly ID, access rights) and structural constraints before being submitted to the Dashboard

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (policy_apply_invoked & yaml_validation_failed) the policy shall immediately satisfy no_dashboard_request_sent
```

## Source
See `specs/system/requirements/SYS-REQ-029.req.yaml` for the machine-readable definition.
