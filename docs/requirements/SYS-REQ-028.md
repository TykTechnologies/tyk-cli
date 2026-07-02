# SYS-REQ-028: policy init scaffolds a new policy YAML file in CLI schema format with sensible defaults given --id and --name flags

<!-- reqproof:req SYS-REQ-028 -->

## Intent
policy init scaffolds a new policy YAML file in CLI schema format with sensible defaults given --id and --name flags

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (policy_init_invoked & policy_init_target_file_exists) the policy shall immediately satisfy policy_exit_bad_args
```

## Source
See `specs/system/requirements/SYS-REQ-028.req.yaml` for the machine-readable definition.
