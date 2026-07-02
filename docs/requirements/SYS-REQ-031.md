# SYS-REQ-031: ParseDuration parses human-friendly duration strings (e.g. 30d, 1h, 24h) into integer seconds; FormatDuration round-trips seconds back to a human-readable string

<!-- reqproof:req SYS-REQ-031 -->

## Intent
ParseDuration parses human-friendly duration strings (e.g. 30d, 1h, 24h) into integer seconds; FormatDuration round-trips seconds back to a human-readable string

## Context
- Level: System
- Component: policy
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when valid_duration_string the policy shall immediately satisfy duration_positive
```

## Source
See `specs/system/requirements/SYS-REQ-031.req.yaml` for the machine-readable definition.
