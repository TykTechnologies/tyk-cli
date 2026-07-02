# SYS-REQ-048: Config.Validate rejects any environment missing name, dashboard_url, auth_token, or org_id, and rejects malformed dashboard URLs; returns a descriptive error citing the environment name

<!-- reqproof:req SYS-REQ-048 -->

## Intent
Config.Validate rejects any environment missing name, dashboard_url, auth_token, or org_id, and rejects malformed dashboard URLs; returns a descriptive error citing the environment name

## Context
- Level: System
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when dashboard_url_empty the config shall immediately satisfy validate_error_returned
```

## Source
See `specs/system/requirements/SYS-REQ-048.req.yaml` for the machine-readable definition.
