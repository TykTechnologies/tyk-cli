# STK-REQ-003: As an API operator I can configure the CLI's connection to one or more Tyk Dashboard environments — via an interactive wizard, named environments, flags, or environment variables — with clear precedence rules so that no secret is ever accidentally hardcoded

<!-- reqproof:req STK-REQ-003 -->

## Intent
As an API operator I can configure the CLI's connection to one or more Tyk Dashboard environments — via an interactive wizard, named environments, flags, or environment variables — with clear precedence rules so that no secret is ever accidentally hardcoded

## Context
- Level: Stakeholder
- Component: config
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Rationale
Operators work across multiple environments (dev, staging, prod). A first-class configuration system prevents credential confusion and makes CI integration safe.

## Source
See `specs/stakeholder/requirements/STK-REQ-003.req.yaml` for the machine-readable definition.
