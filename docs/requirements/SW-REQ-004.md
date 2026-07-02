# SW-REQ-004: ParseDuration returns a non-nil error for empty, negative, fractional, or unsupported-suffix inputs

<!-- reqproof:req SW-REQ-004 -->

## Intent
ParseDuration returns a non-nil error for empty, negative, fractional, or unsupported-suffix inputs

## Context
- Level: Software
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: SYS-REQ-031

## Formalization
```
when parse_duration_invalid_input the policy shall immediately satisfy parse_duration_error
```

## Source
See `specs/software/requirements/SW-REQ-004.req.yaml` for the machine-readable definition.
