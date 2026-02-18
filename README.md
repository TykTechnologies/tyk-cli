# Tyk CLI

A powerful command-line interface for managing Tyk APIs and configurations. Built to streamline API lifecycle management with OpenAPI Specification (OAS) support.

> 📖 **[View Complete Documentation](https://sedkis.github.io/tyk-cli/)** | 🚀 **[Get Started](https://sedkis.github.io/tyk-cli/getting-started)** | 💡 **[Examples](https://sedkis.github.io/tyk-cli/examples/)**

[![Go Version](https://img.shields.io/github/go-mod/go-version/sedkis/tyk-cli?style=flat-square)](https://golang.org/)
[![License](https://img.shields.io/github/license/sedkis/tyk-cli?style=flat-square)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/sedkis/tyk-cli?style=flat-square)](https://github.com/sedkis/tyk-cli/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/sedkis/tyk-cli/total?style=flat-square)](https://github.com/sedkis/tyk-cli/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/sedkis/tyk-cli?style=flat-square)](https://goreportcard.com/report/github.com/sedkis/tyk-cli)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen?style=flat-square)](https://sedkis.github.io/tyk-cli/)

## ✨ Features

- 🚀 **Interactive Setup Wizard** - Get started quickly with guided configuration
- 🌍 **Unified Config/Environment System** - Named environments with seamless switching
- 📝 **OpenAPI First** - Native support for OAS 3.0 specifications  
- 🔧 **Flexible Configuration** - Environment variables, unified config file, or CLI flags
- 🎨 **Beautiful CLI** - Colorful, intuitive command-line experience
- ✅ **Comprehensive Testing** - >80% test coverage with live environment validation

## Licensed Tyk only
- This CLI is designed to work with the licensed version of Tyk - via Control Plane.

## 🚀 Quick Start

### Installation

#### Homebrew (Recommended)

```bash
# Add the Tyk tap
brew tap sedkis/tyk

# Install the CLI
brew install tyk

## Check version
$ tyk -v
tyk version 0.2.1

## UPGRADING the cli
$ brew update

$ brew upgrade tyk

## Check version
$ tyk -v
tyk version 0.2.1
```

#### Direct Download

```bash
# Download and install the latest release
curl -L "https://github.com/sedkis/tyk-cli/releases/latest/download/tyk-cli_$(uname -s)_$(uname -m).tar.gz" | tar xz
sudo mv tyk /usr/local/bin/

# Verify installation
tyk --version
```

#### From Source

```bash
git clone https://github.com/sedkis/tyk-cli.git
cd tyk-cli
go build -o tyk .
sudo mv tyk /usr/local/bin/
```

### Initialize Configuration

```bash
# Run the interactive setup wizard
tyk init
```

> 💡 **Need more help?** Check out our [complete documentation](https://sedkis.github.io/tyk-cli/) with detailed guides, examples, and troubleshooting tips!

## 📋 Commands

### Configuration/Environment Management
```bash
# Interactive setup wizard
tyk init                    # Full setup wizard with multiple environments

# Environment management
tyk config list            # List all environments
tyk config use             # Switch environnment interactively
tyk config use staging     # Switch to staging environment
tyk config current         # Show current environment
tyk config set dashboard-url https://api.tyk.io  # Update current environment
```

### API Management
```bash
# Quick API Creation
tyk api create --name "User Service" --upstream-url https://users.api.com

# Create from OpenAPI Spec Management
tyk api import-oas --file petstore.yaml           # Import external OpenAPI spec
tyk api import-oas --url https://api.example.com/openapi.json  # Import from URL
tyk api update-oas <api-id> --file new-spec.yaml  # Update API's OpenAPI spec only

# Tyk-Enhanced OAS Management (GitOps)
# If the file contains x-tyk-api-gateway.info.id, apply will upsert:
# update if it exists, or create with the same ID if missing
tyk api apply --file enhanced-api.yaml            # Idempotent upsert (update or create)

# General Operations
tyk api list                        # List all APIs
tyk api list -i                     # Interactive
tyk api get <api-id>                               # Get API details
tyk api get <api-id> --oas-only                   # Get OpenAPI spec only
tyk api delete <api-id>             # Delete API (with confirmation)
tyk api delete <api-id> --yes       # Delete without confirmation

# Utilities (Phase 3)
tyk api convert --file api.yaml --format apidef  # Convert OAS to Tyk format
```

### Policy Management
```bash
# Scaffold a new policy file
tyk policy init --id gold --name "Gold Plan"

# List policies
tyk policy list                     # Paginated table
tyk policy list --json              # JSON output

# Get a policy (outputs CLI-schema YAML)
tyk policy get <policy-id>          # YAML to stdout, summary to stderr
tyk policy get <policy-id> --json   # JSON output

# Apply a policy (idempotent upsert — creates or updates)
tyk policy apply -f policy.yaml     # From file
tyk policy apply -f -               # From stdin

# Delete a policy
tyk policy delete <policy-id>       # Interactive confirmation
tyk policy delete <policy-id> --yes # Skip confirmation
```

Policies use a human-friendly **CLI schema** with readable durations (`30d`, `1h`) and API selectors by name, listen path, or tags. The CLI converts to Dashboard wire format on apply. See the [Policy Guide](https://sedkis.github.io/tyk-cli/manage-policies) for the full YAML reference.

## ⚙️ Configuration

The Tyk CLI uses a unified environment/configuration system with the following precedence (highest to lowest):

1. **Command-line flags** (`--dash-url`, `--auth-token`, `--org-id`)
2. **Environment variables** (`TYK_DASH_URL`, `TYK_AUTH_TOKEN`, `TYK_ORG_ID`)
3. **Named environments in config file** (`~/.config/tyk/cli.toml`)

Each "environment" is simply a named set of configuration values.

### Environment Variables

```bash
export TYK_DASH_URL=http://localhost:3000
export TYK_AUTH_TOKEN=your-api-token
export TYK_ORG_ID=your-org-id
```

### Config File (Unified Environment System)

Configuration is automatically saved to `~/.config/tyk/cli.toml`:

```toml
default_environment = "dev"

[environments.dev]
dashboard_url = "http://localhost:3000"
auth_token = "dev-api-token" 
org_id = "dev-org-id"

[environments.staging]
dashboard_url = "https://staging.tyk.io"
auth_token = "staging-token"
org_id = "staging-org-id"

[environments.production]
dashboard_url = "https://api.yourcompany.com"
auth_token = "prod-token"
org_id = "prod-org-id"
```

## 🔍 Finding Your Credentials

### Dashboard URL
- **Local Development**: `http://localhost:3000` (default)
- **Tyk Cloud**: `https://admin.cloud.tyk.io`
- **Self-hosted**: Your custom domain

### Auth Token
1. Log into your Tyk Dashboard
2. Go to **Users** → Your User Profile
3. Find **API Access Credentials**
4. Copy the **Auth Token**

### Organization ID
- Found next to your API Token, in User Profile.

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Make (optional)

### Setup

```bash
# Clone the repository
git clone https://github.com/sedkis/tyk-cli.git
cd tyk-cli

# Install dependencies
go mod download

# Build the CLI
go build -o tyk .

# Run tests
go test ./...

# Build with make (if available)
make build
make test
```

### Project Structure

```
tyk-cli/
├── cmd/              # Entry point
├── internal/         # Internal packages
│   ├── cli/          # Cobra command definitions (api, policy, config)
│   ├── config/       # Configuration management
│   ├── client/       # HTTP client for Tyk Dashboard API
│   ├── policy/       # Policy domain logic (duration, validation, selectors, conversion)
│   ├── oas/          # OAS file handling
│   └── filehandler/  # File utilities
├── pkg/types/        # Shared types (API, policy, config)
├── test/             # Integration tests
└── docs/             # Documentation (Jekyll site)
```

## 🗺️ Roadmap

### 🔧 Other lifecycle objects
- Tyk API Tokens / Credentials

### 🔧 Enhanced Features
- `tyk api convert` - Convert between OAS and Tyk API definition formats
- Enhanced error handling and user experience improvements
- Advanced JSON output formatting
- API versioning commands (`versions list/create/switch-default`)
- API validation and linting
- GitOps diff functionality

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Common Development Tasks

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint code (if golangci-lint is installed)
golangci-lint run

# Format code
go fmt ./...
```

## 📄 License

This project is licensed under the [MIT License](LICENSE).

## 🆘 Support

- 📖 **Documentation**: [Complete Documentation](https://sedkis.github.io/tyk-cli/)
- 🚀 **Getting Started**: [Installation & Setup Guide](https://sedkis.github.io/tyk-cli/getting-started)
- 🐛 **Issues**: [GitHub Issues](https://github.com/sedkis/tyk-cli/issues)
- 💬 **Community**: [Tyk Community Forum](https://community.tyk.io/)

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Survey](https://github.com/AlecAivazis/survey) - Interactive prompts
- [Color](https://github.com/fatih/color) - Terminal colors
