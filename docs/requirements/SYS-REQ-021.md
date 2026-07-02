# SYS-REQ-021: api commands communicate with the Tyk Dashboard via an authenticated HTTP client that sets Authorization and x-org-id headers on every request

<!-- reqproof:req SYS-REQ-021 -->

## Intent
api commands communicate with the Tyk Dashboard via an authenticated HTTP client that sets Authorization and x-org-id headers on every request

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when dashboard_request_sent the api shall immediately satisfy (authorization_header_set & org_id_header_set)
```

## Source
See `specs/system/requirements/SYS-REQ-021.req.yaml` for the machine-readable definition.
