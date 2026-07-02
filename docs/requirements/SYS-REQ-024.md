# SYS-REQ-024: policy list returns paginated policies from the Dashboard; page defaults to 1; supports --json flag

<!-- reqproof:req SYS-REQ-024 -->

## Intent
policy list returns paginated policies from the Dashboard; page defaults to 1; supports --json flag

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when policy_page_flag_omitted the policy shall immediately satisfy policy_page_equals_one
```

## Source
See `specs/system/requirements/SYS-REQ-024.req.yaml` for the machine-readable definition.
