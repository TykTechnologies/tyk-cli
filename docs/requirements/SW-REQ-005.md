# SW-REQ-005: 'ParseDuration converts a suffixed input using the canonical multiplier table: s=1, m=60, h=3600, d=86400; the resolved value is positive'

<!-- reqproof:req SW-REQ-005 -->

## Intent
'ParseDuration converts a suffixed input using the canonical multiplier table: s=1, m=60, h=3600, d=86400; the resolved value is positive'

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
when parse_duration_with_suffix the policy shall immediately satisfy (resolved_seconds_matches_suffix_table & duration_positive)
```

## Source
See `specs/software/requirements/SW-REQ-005.req.yaml` for the machine-readable definition.
