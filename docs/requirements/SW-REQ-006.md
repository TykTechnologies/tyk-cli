# SW-REQ-006: 'ParseDuration and FormatDuration roundtrip: ParseDuration(FormatDuration(n)) == n for any non-negative n expressible as an exact unit multiple'

<!-- reqproof:req SW-REQ-006 -->

## Intent
'ParseDuration and FormatDuration roundtrip: ParseDuration(FormatDuration(n)) == n for any non-negative n expressible as an exact unit multiple'

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
when valid_seconds_input the policy shall immediately satisfy format_parse_roundtrip_equal
```

## Source
See `specs/software/requirements/SW-REQ-006.req.yaml` for the machine-readable definition.
