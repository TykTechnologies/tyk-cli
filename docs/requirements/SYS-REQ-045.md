# SYS-REQ-045: tyk config set updates dashboard_url, auth_token, and/or org_id on the active environment; refuses with exit 2 when no flags are supplied

<!-- reqproof:req SYS-REQ-045 -->

## Intent
tyk config set updates dashboard_url, auth_token, and/or org_id on the active environment; refuses with exit 2 when no flags are supplied

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: STK-REQ-003

## Formalization
```
when set_invoked_with_no_flags the config shall immediately satisfy set_no_flags_rejected_with_bad_args
```

## Source
See `specs/system/requirements/SYS-REQ-045.req.yaml` for the machine-readable definition.
