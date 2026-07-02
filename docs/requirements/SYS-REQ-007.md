# SYS-REQ-007: api delete removes an API by ID; prompts for confirmation unless --yes is passed; exits with code 3 if the API does not exist

<!-- reqproof:req SYS-REQ-007 -->

## Intent
api delete removes an API by ID; prompts for confirmation unless --yes is passed; exits with code 3 if the API does not exist

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when (delete_invoked & delete_target_does_not_exist) the api shall immediately satisfy api_exit_not_found
```

## Source
See `specs/system/requirements/SYS-REQ-007.req.yaml` for the machine-readable definition.
