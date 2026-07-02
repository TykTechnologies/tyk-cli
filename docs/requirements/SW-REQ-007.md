# SW-REQ-007: GenerateListenPath always returns a value that starts with '/' and ends with '/' and is valid as a listen path

<!-- reqproof:req SW-REQ-007 -->

## Intent
GenerateListenPath always returns a value that starts with '/' and ends with '/' and is valid as a listen path

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
when listen_path_generated the api shall immediately satisfy (result_has_slash_prefix_and_suffix & listen_path_valid)
```

## Source
See `specs/software/requirements/SW-REQ-007.req.yaml` for the machine-readable definition.
