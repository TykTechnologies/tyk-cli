# Technology Stack: policies-mgmt

**Feature**: Security Policy Management for Tyk CLI
**Wave**: DESIGN
**Date**: 2026-02-18

---

## Dependency Analysis

### Reused Dependencies (No Changes to go.mod)

| Dependency | License | Version | Usage in Policy Module |
|---|---|---|---|
| `github.com/spf13/cobra` | Apache-2.0 | v1.10.1 | Command tree for `tyk policy` subcommands |
| `github.com/spf13/viper` | MIT | v1.20.1 | Config loading (via existing root.go) |
| `github.com/fatih/color` | MIT | v1.18.0 | Colored output for policy summaries and tables |
| `github.com/AlecAivazis/survey/v2` | MIT | v2.3.7 | Interactive prompts in `policy init` |
| `github.com/stretchr/testify` | MIT | v1.11.1 | Assertions in policy unit tests |
| `gopkg.in/yaml.v3` | Apache-2.0 | v3.0.1 | YAML marshal/unmarshal for policy files |
| `golang.org/x/term` | BSD-3-Clause | v0.0.0 | Terminal detection for interactive mode |

### New Dependencies

**None.**

All policy functionality is implemented using the Go standard library and existing dependencies. Specifically:

| Capability | Approach | Why No New Dep |
|---|---|---|
| Duration parsing (`30d` -> seconds) | Custom parser (~30 lines) using `strconv` and string suffix matching | Go `time.ParseDuration` lacks `d` (days); external libs add overhead for trivial logic |
| Fuzzy string matching (selector suggestions) | Levenshtein distance (~20 lines) using standard library | Operating on small datasets (<1000 API names); a library dependency for 20 lines is not justified |
| Schema validation | Struct tag validation + manual field checks | The schema has ~10 fields; a validation framework (e.g., `go-playground/validator`) adds 3 transitive deps for minimal gain |
| YAML with comments (scaffold) | `text/template` from stdlib + `yaml.v3` marshal | Comment-annotated scaffold is a template string, not runtime YAML manipulation |

### Dependency Decision Criteria

1. **Does the existing go.mod already include it?** -> Reuse
2. **Is the implementation < 50 lines of straightforward code?** -> Standard library
3. **Does it add transitive dependencies?** -> Strong bias against
4. **Is it well-maintained (commits in last 6 months, >500 stars)?** -> Required if adding

### When to Revisit (Phase 2+)

- If fuzzy matching needs to support CJK characters or phonetic similarity -> consider `github.com/lithammer/fuzzysearch` (MIT, 1.3k stars)
- If `policy init` scaffold needs rich template features -> consider `text/template` functions (stdlib, no dep change)
- If interactive `policy list` needs TUI beyond current arrow-key navigation -> consider `github.com/charmbracelet/bubbletea` (MIT, 26k stars)

---

## Build and Test Infrastructure

### Existing (Reused As-Is)

| Tool | Purpose |
|---|---|
| `go test ./...` | Unit tests with testify assertions |
| `net/http/httptest` | Mock Dashboard server in client tests |
| `go build -ldflags` | Binary compilation with version info |

### No Changes to Build Pipeline

The policy module follows the same file conventions and package layout. No new build steps, no code generation, no additional tooling.

---

## Go Version Compatibility

The project uses `go 1.24.4` (per `go.mod`). All policy module code uses standard library features available since Go 1.21+. No version-specific features required.
