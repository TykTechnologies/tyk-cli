# SYS-REQ-026: 'policy apply performs an idempotent upsert: creates a new policy or updates an existing one based on the id field in the YAML file; accepts --file or stdin (-)'

<!-- reqproof:req SYS-REQ-026 -->

## Intent
'policy apply performs an idempotent upsert: creates a new policy or updates an existing one based on the id field in the YAML file; accepts --file or stdin (-)'

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (policy_apply_invoked & policy_target_exists) the policy shall immediately satisfy policy_apply_takes_update_path
```

## Source
See `specs/system/requirements/SYS-REQ-026.req.yaml` for the machine-readable definition.
