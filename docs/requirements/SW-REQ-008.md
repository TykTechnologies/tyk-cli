# SW-REQ-008: GenerateListenPath collapses runs of non-alphanumeric characters in the title into a single '-' in the slug

<!-- reqproof:req SW-REQ-008 -->

## Intent
GenerateListenPath collapses runs of non-alphanumeric characters in the title into a single '-' in the slug

## Context
- Level: Software
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: SYS-REQ-010

## Formalization
```
when title_has_non_alphanumeric_chars the api shall immediately satisfy result_collapses_non_alphanumeric_to_hyphen
```

## Source
See `specs/software/requirements/SW-REQ-008.req.yaml` for the machine-readable definition.
