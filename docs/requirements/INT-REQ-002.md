# INT-REQ-002: The Dashboard error taxonomy shall be mapped to exit codes via a single typed classifier (classifyDashboardError). Every api and policy subcommand shall route Dashboard errors through the classifier so that shell callers observe a stable status-class-to-exit-code contract (401→5, 403→6, 404→3, 409→4, 429→7, 5xx→8).

<!-- reqproof:req INT-REQ-002 -->

## Intent
The Dashboard error taxonomy shall be mapped to exit codes via a single typed classifier (classifyDashboardError). Every api and policy subcommand shall route Dashboard errors through the classifier so that shell callers observe a stable status-class-to-exit-code contract (401→5, 403→6, 404→3, 409→4, 429→7, 5xx→8).

## Context
- Level: Interface
- Component: cli_dashboard_contract
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: SYS-REQ-013

## Formalization
```
when dashboard_returned_typed_error_response the cli_dashboard_contract shall immediately satisfy exit_code_matches_status_class
```

## Source
See `specs/interface/requirements/INT-REQ-002.req.yaml` for the machine-readable definition.

documents INT-REQ-002
