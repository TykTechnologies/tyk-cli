# Journey: OAS Version Management

## Emotional Arc

```
Confidence
    ^
    |                                          **** SWITCH DEFAULT
    |                                     ****      (relieved,
    |                              ****              in control)
    |                        ****
    |                   **** CREATE VERSION
    |              ****      (productive,
    |         ****            building)
    |    ****
    |**** LIST VERSIONS          GET VERSION
    |     (oriented,              (informed,
    |      grounded)               ready)
    +-------------------------------------------------> Time
```

## Step 1: List Versions

**Trigger**: Developer needs to understand current version landscape before making changes.
**Feeling**: Oriented, grounded -- "I can see what exists."

```
$ tyk api versions list --api-id abc123

  API: Petstore (abc123)

  VERSION    DEFAULT
  v1         *
  v2
  v3

  3 versions found. Default: v1
```

```
$ tyk api versions list --api-id abc123 --json

{
  "api_id": "abc123",
  "api_name": "Petstore",
  "default": "v1",
  "versions": ["v1", "v2", "v3"]
}
```

**Error: API not found**
```
$ tyk api versions list --api-id nonexistent

  Error: API not found: nonexistent
  $ echo $?
  3
```

---

## Step 2: Get Specific Version

**Trigger**: Developer needs to inspect a version's OAS document before creating or modifying.
**Feeling**: Informed, ready -- "I know exactly what this version contains."

```
$ tyk api get --api-id abc123 --version-name v2

  API: Petstore (abc123)
  Version: v2
  Listen Path: /petstore-v2/
  Upstream: https://api.petstore.io/v2
  ---
  openapi: "3.0.3"
  info:
    title: "Petstore"
    version: "2.0.0"
  ...
```

```
$ tyk api get --api-id abc123 --version-name v2 --json
{
  "openapi": "3.0.3",
  "info": { "title": "Petstore", "version": "2.0.0" },
  "x-tyk-api-gateway": { ... }
}
```

**Error: Version not found**
```
$ tyk api get --api-id abc123 --version-name v99

  Error: version "v99" not found for API abc123
  Available versions: v1, v2, v3
  $ echo $?
  3
```

---

## Step 3: Create a New Version

**Trigger**: Developer has a new OAS spec and wants to add it as a version to an existing API.
**Feeling**: Productive, building -- "I'm extending my API safely."

```
$ tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3

  Version created successfully.

  API:          Petstore (abc123)
  Version:      v3
  Listen Path:  /petstore-v3/
  Upstream:     https://api.petstore.io/v3
  Default:      no

  To make this the default version:
    tyk api versions switch-default --api-id abc123 --version-name v3
```

```
$ tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3 --set-default

  Version created successfully.

  API:          Petstore (abc123)
  Version:      v3 (default)
  Listen Path:  /petstore-v3/
  Upstream:     https://api.petstore.io/v3
  Default:      yes
```

```
$ tyk api versions create --api-id abc123 -f petstore-v3.yaml --version-name v3 --json
{
  "action": "created",
  "api_id": "abc123",
  "version_name": "v3",
  "listen_path": "/petstore-v3/",
  "upstream_url": "https://api.petstore.io/v3",
  "is_default": false
}
```

**Error: Version name conflict**
```
$ tyk api versions create --api-id abc123 -f petstore-v2.yaml --version-name v2

  Error: version "v2" already exists for API abc123
  Use 'tyk api apply -f petstore-v2.yaml' to update an existing version.
  $ echo $?
  4
```

**Error: Invalid OAS document**
```
$ tyk api versions create --api-id abc123 -f broken.yaml --version-name v4

  Error: invalid OAS document: missing info section
  $ echo $?
  2
```

---

## Step 4: Switch Default Version

**Trigger**: Developer wants to route default traffic to a different version.
**Feeling**: Relieved, in control -- "Traffic is going where I want it."

```
$ tyk api versions switch-default --api-id abc123 --version-name v3

  Default version switched.

  API:              Petstore (abc123)
  Previous default: v1
  New default:      v3
```

```
$ tyk api versions switch-default --api-id abc123 --version-name v3 --json
{
  "action": "switched_default",
  "api_id": "abc123",
  "previous_default": "v1",
  "new_default": "v3"
}
```

**Error: Version does not exist**
```
$ tyk api versions switch-default --api-id abc123 --version-name v99

  Error: version "v99" not found for API abc123
  Available versions: v1, v2, v3
  $ echo $?
  3
```

**Error: Already the default**
```
$ tyk api versions switch-default --api-id abc123 --version-name v1

  Version "v1" is already the default for API abc123.
  $ echo $?
  0
```

---

## Step 5: Apply Integration (Declarative/GitOps)

**Trigger**: Developer manages versioned APIs through file-based workflows.
**Feeling**: Confident, systematic -- "My version config is in source control."

### Single file apply with version context
```
$ tyk api apply -f petstore-v3.yaml --version-name v3 --base-api-id abc123

  Applied successfully.

  API:       Petstore (abc123)
  Version:   v3 (created)
  Operation: create_version
```

### Batch apply recognizes versioned files
```
apis/
  petstore-base.yaml       # base API (has x-tyk-api-gateway with id)
  petstore-v2.yaml         # version file (has x-tyk-version-of + version-name)
  petstore-v3.yaml         # version file

$ tyk apply -f apis/

  [1/3] petstore-base.yaml ... OK (updated)
  [2/3] petstore-v2.yaml ... OK (version created: v2)
  [3/3] petstore-v3.yaml ... OK (version created: v3)

  Apply complete: 3/3 succeeded
```

### Version metadata in OAS files (convention)
Files that represent versions of an existing API use an extension:
```yaml
x-tyk-api-gateway:
  info:
    name: "Petstore"
    versioning:
      base_api_id: "abc123"
      version_name: "v3"
      set_default: false
```

---

## Integration Flow

```
                    +------------------+
                    | List Versions    |
                    | (orientation)    |
                    +--------+---------+
                             |
              +--------------+--------------+
              |                             |
    +---------v----------+       +----------v---------+
    | Get Version        |       | Create Version     |
    | (inspect before    |       | (from OAS file)    |
    | acting)            |       +----------+---------+
    +--------------------+                  |
                                            |
                              +-------------v-----------+
                              | Switch Default          |
                              | (route traffic)         |
                              +-------------+-----------+
                                            |
                              +-------------v-----------+
                              | Apply (declarative)     |
                              | (GitOps integration)    |
                              +-------------------------+
```
