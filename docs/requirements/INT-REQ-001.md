# INT-REQ-001: All Dashboard HTTP requests issued by any tyk-cli subcommand shall carry the Authorization and x-org-id headers established by the client.NewClient contract. This interface applies across api, policy, and config components.

<!-- reqproof:req INT-REQ-001 -->

## Intent
All Dashboard HTTP requests issued by any tyk-cli subcommand shall carry the Authorization and x-org-id headers established by the client.NewClient contract. This interface applies across api, policy, and config components.

## Context
- Level: Interface
- Component: cli_dashboard_contract
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: SYS-REQ-021

## Formalization
```
when dashboard_http_request_issued the cli_dashboard_contract shall immediately satisfy authorization_and_org_headers_present
```

## Source
See `specs/interface/requirements/INT-REQ-001.req.yaml` for the machine-readable definition.

documents INT-REQ-001
