---
title: Getting Started
nav_order: 2
---

# Getting Started

Welcome! You’ll be productive in ~3 minutes.

## 1) Install

macOS, using brew
  ```
  brew tap sedkis/tyk && brew install tyk
  ```
Or Linux/macOS (tarball)
  ```
  curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
  sudo mv tyk /usr/local/bin/
  ```

2) Say hi
```
tyk --help
```

## 3) Create your first environment config
```
tyk init
## or
tyk config add dev --dashboard-url http://localhost:3000 --auth-token dev-token --org-id dev-org

## then
tyk config use dev
```

## 4) Create your first API
Create from scratch:
```bash
tyk api create --name httpbin --upstream-url http://httpbingo.org
```

Response:
```bash
✓ API created successfully!
  API ID:         8acf2c7c0d6d4bf3707b429afeaed791
  Name:           httpbin
  Version:        v1
  Listen Path:    /httpbin/
  Upstream URL:   http://httpbingo.org
  Default Version: v1

Next steps:
  tyk api get 8acf2c7c0d6d4bf3707b429afeaed791                           # View full configuration
  tyk api get 8acf2c7c0d6d4bf3707b429afeaed791 --oas-only > api.yaml    # Export for editing
```

## OAS Workflows
```
tyk api import-oas --file path/to/my-api.yaml
```

## 5) Create your first policy

Generate a scaffold:
```bash
tyk policy init --id gold --name "Gold Plan"
```

Edit `policies/gold.yaml` to configure rate limits and API access:
```yaml
apiVersion: tyk.tyktech/v1
kind: Policy
metadata:
  id: gold
  name: Gold Plan
spec:
  rateLimit:
    requests: 1000
    per: 1m
  quota:
    limit: 100000
    period: 30d
  access:
    - name: httpbin          # Resolves by API name
      versions: [Default]
```

Apply it:
```bash
tyk policy apply -f policies/gold.yaml
```

Verify:
```bash
tyk policy list
tyk policy get gold
```

> See the full [Policy Guide]({{ site.baseurl }}/manage-policies) for selectors, durations, and advanced usage.

## 6) Check it worked
- See your API and policy in the Tyk Dashboard
- Hit a simple endpoint or health route

Tips
- Keep tokens out of shell history by using env vars; see {{ site.baseurl }}/configuration
- Start with a small OAS to keep feedback tight
- Use --dry-run if you want a no-changes preview (when available)
