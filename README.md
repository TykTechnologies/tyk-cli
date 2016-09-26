# tyk-cli

Tyk CLI utility.

## Available modules

### Bundle

This module provides useful commands for working with custom middleware bundles. The most basic command is `build`:

Assuming you're on a directory that contains your required bundle files and a **bundle manifest**, you could run:

```
tyk-cli bundle build -output bundle-latest.zip
```

The bundle will contain a `manifest.json` with the computed checksum and signature.
