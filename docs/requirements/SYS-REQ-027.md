# SYS-REQ-027: policy delete removes a policy by ID with an interactive confirmation prompt unless --yes is passed; exits with code 3 if the policy does not exist

<!-- reqproof:req SYS-REQ-027 -->

## Intent
policy delete removes a policy by ID with an interactive confirmation prompt unless --yes is passed; exits with code 3 if the policy does not exist

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (policy_delete_invoked & policy_delete_target_does_not_exist) the policy shall immediately satisfy policy_exit_not_found
```

## Source
See `specs/system/requirements/SYS-REQ-027.req.yaml` for the machine-readable definition.
