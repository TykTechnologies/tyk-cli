# SW-REQ-002: FuzzySuggestions returns candidates sorted ascending by Levenshtein distance to the query

<!-- reqproof:req SW-REQ-002 -->

## Intent
FuzzySuggestions returns candidates sorted ascending by Levenshtein distance to the query

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
when fuzzy_invoked the policy shall immediately satisfy fuzzy_result_sorted_ascending
```

## Source
See `specs/software/requirements/SW-REQ-002.req.yaml` for the machine-readable definition.
