# SYS-REQ-032: ResolveAccessEntries resolves policy access entries by API name, listen path, or tag against the live API list; returns an error for any selector that matches zero or more than one API

<!-- reqproof:req SYS-REQ-032 -->

## Intent
ResolveAccessEntries resolves policy access entries by API name, listen path, or tag against the live API list; returns an error for any selector that matches zero or more than one API

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when ambiguous_selector the policy shall immediately satisfy resolve_error_returned
```

## Source
See `specs/system/requirements/SYS-REQ-032.req.yaml` for the machine-readable definition.
