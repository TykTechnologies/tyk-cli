# SYS-REQ-034: All policy subcommands exit 2 on invalid arguments or missing required flags

<!-- reqproof:req SYS-REQ-034 -->

## Intent
All policy subcommands exit 2 on invalid arguments or missing required flags

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when policy_invalid_args the policy shall immediately satisfy policy_exit_bad_args
```

## Source
See `specs/system/requirements/SYS-REQ-034.req.yaml` for the machine-readable definition.
