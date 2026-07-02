# SYS-REQ-033: All policy subcommands exit 0 on success

<!-- reqproof:req SYS-REQ-033 -->

## Intent
All policy subcommands exit 0 on success

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when policy_operation_succeeded the policy shall immediately satisfy policy_exit_zero
```

## Source
See `specs/system/requirements/SYS-REQ-033.req.yaml` for the machine-readable definition.
