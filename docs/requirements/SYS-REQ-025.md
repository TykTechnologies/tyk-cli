# SYS-REQ-025: policy get retrieves a single policy by ID and outputs CLI-schema YAML to stdout with a summary to stderr; supports --json flag

<!-- reqproof:req SYS-REQ-025 -->

## Intent
policy get retrieves a single policy by ID and outputs CLI-schema YAML to stdout with a summary to stderr; supports --json flag

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when json_flag_set the policy shall immediately satisfy output_is_json
```

## Source
See `specs/system/requirements/SYS-REQ-025.req.yaml` for the machine-readable definition.
