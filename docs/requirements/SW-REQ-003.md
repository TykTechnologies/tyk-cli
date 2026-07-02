# SW-REQ-003: FuzzySuggestions returns an empty slice when invoked with an empty candidate list

<!-- reqproof:req SW-REQ-003 -->

## Intent
FuzzySuggestions returns an empty slice when invoked with an empty candidate list

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
when fuzzy_invoked_with_empty_list the policy shall immediately satisfy (fuzzy_result_is_empty & resolve_error_returned)
```

## Source
See `specs/software/requirements/SW-REQ-003.req.yaml` for the machine-readable definition.
