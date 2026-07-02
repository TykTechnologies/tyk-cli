# SYS-REQ-047: Environments may declare a positive timeout_seconds; when set, the HTTP client uses that timeout instead of the 30s default, allowing operators to extend the deadline for slow Dashboards

<!-- reqproof:req SYS-REQ-047 -->

## Intent
Environments may declare a positive timeout_seconds; when set, the HTTP client uses that timeout instead of the 30s default, allowing operators to extend the deadline for slow Dashboards

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
when timeout_seconds_set_on_environment the config shall immediately satisfy http_client_uses_environment_timeout
```

## Source
See `specs/system/requirements/SYS-REQ-047.req.yaml` for the machine-readable definition.
