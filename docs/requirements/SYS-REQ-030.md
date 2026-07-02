# SYS-REQ-030: policy documents are bi-directionally converted between CLI YAML (PolicyFile) and Dashboard wire JSON (DashboardPolicy) preserving all access rights, duration, and metadata

<!-- reqproof:req SYS-REQ-030 -->

## Intent
policy documents are bi-directionally converted between CLI YAML (PolicyFile) and Dashboard wire JSON (DashboardPolicy) preserving all access rights, duration, and metadata

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when valid_cli_policy_input the policy shall immediately satisfy wire_roundtrip_preserves_fields
```

## Source
See `specs/system/requirements/SYS-REQ-030.req.yaml` for the machine-readable definition.
