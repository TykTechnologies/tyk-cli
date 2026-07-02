# SW-REQ-001: FuzzySuggestions returns a slice no longer than the requested max (n argument)

<!-- reqproof:req SW-REQ-001 -->

## Intent
FuzzySuggestions returns a slice no longer than the requested max (n argument)

## Context
- Level: Software
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: SYS-REQ-032

## Formalization
```
when fuzzy_invoked the policy shall immediately satisfy fuzzy_result_count_bounded
```

## Source
See `specs/software/requirements/SW-REQ-001.req.yaml` for the machine-readable definition.
