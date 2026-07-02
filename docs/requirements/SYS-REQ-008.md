# SYS-REQ-008: stripExistingAPIID removes x-tyk-api-gateway.info.id from any OAS document before it is sent to the create endpoint, ensuring the Dashboard always generates a new API ID

<!-- reqproof:req SYS-REQ-008 -->

## Intent
stripExistingAPIID removes x-tyk-api-gateway.info.id from any OAS document before it is sent to the create endpoint, ensuring the Dashboard always generates a new API ID

## Context
- Level: System
- Component: api
- Type: guarantee
- Category: functional
- Status: approved
- Assurance: B

## Formalization
```
when id_present the api shall immediately satisfy id_stripped
```

## Source
See `specs/system/requirements/SYS-REQ-008.req.yaml` for the machine-readable definition.
