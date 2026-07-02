# INT-REQ-003: Every OAS submission path (api import-oas, api update-oas) shall invoke oas.ValidateOASStructure before any Dashboard round-trip. Malformed OAS documents must be rejected locally with ExitBadArgs so that operators get fast feedback and the Dashboard is not asked to reject known-bad input.

<!-- reqproof:req INT-REQ-003 -->

## Intent
Every OAS submission path (api import-oas, api update-oas) shall invoke oas.ValidateOASStructure before any Dashboard round-trip. Malformed OAS documents must be rejected locally with ExitBadArgs so that operators get fast feedback and the Dashboard is not asked to reject known-bad input.

## Context
- Level: Interface
- Component: cli_dashboard_contract
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B
- Parent: SYS-REQ-012

## Formalization
```
when oas_document_submitted_to_dashboard the cli_dashboard_contract shall immediately satisfy oas_passes_structural_validation
```

## Source
See `specs/interface/requirements/INT-REQ-003.req.yaml` for the machine-readable definition.

documents INT-REQ-003
