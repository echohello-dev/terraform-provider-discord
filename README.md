# Terraform Provider for Discord

This Terraform provider allows you to manage Discord resources such as servers (guilds), channels, and roles using infrastructure-as-code principles.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.9
- [Go](https://golang.org/doc/install) >= 1.23
- [mise](https://mise.jdx.dev/) >= 2024.1.0 (for task running and tool management)
- A Discord bot token with appropriate permissions

## Quick Start

Install dependencies with mise:

```bash
mise install
```

## Authentication

The provider requires a Discord bot token. You can provide it either:

- Via the `token` attribute in the provider configuration
- Via the `DISCORD_TOKEN` environment variable

## Development

### Build the provider

```bash
mise run build
```

### Install the provider locally (uses `go install`)

```bash
mise run install
```

### Set up Terraform dev overrides

To test the provider locally without publishing it, configure Terraform to use your locally built binary:

```bash
mise run dev-setup
# Edit ~/.terraformrc and replace the empty path with your GOBIN directory:
# go env GOBIN
```

Alternatively, create `~/.terraformrc` manually:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/echohello-dev/discord" = "/Users/YOU/go/bin"
  }
  direct {}
}
```

> **Note:** When using `dev_overrides`, skip `terraform init`. The provider is loaded directly from the local binary.

### Running tests

```bash
# Unit tests
mise run test

# Acceptance tests (requires DISCORD_TOKEN)
mise run testacc

# Run specific tests
mise run test "TestAccResource_server"
```

### Formatting and linting

```bash
mise run fmt
mise run vet
mise run lint
```

### Generate documentation

```bash
mise run docs
```

Or via `go generate`:

```bash
go generate ./...
```

### Tidy modules

```bash
mise run tidy
```

## Usage

```hcl
terraform {
  required_providers {
    discord = {
      source  = "echohello-dev/discord"
      version = "0.1.0"
    }
  }
}

provider "discord" {
  # token = "YOUR_BOT_TOKEN"
  # Or use DISCORD_TOKEN environment variable
}

resource "discord_server" "example" {
  name = "My Server"
}
```

See the `examples/` directory for more complete usage examples.

## Resources

| Resource | Description |
|----------|-------------|
| `discord_server` | Manage a Discord server (guild) |
| `discord_channel` | Manage a channel within a server |
| `discord_role` | Manage a role within a server |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `discord_server` | Read information about an existing server |

## Available Tasks

Run `mise run` (or `mise run h`) to see all available tasks.

| Task | Description |
|------|-------------|
| `build` | Build the provider binary to `.build/` |
| `install` | Install the provider to `GOBIN` |
| `clean` | Remove built artifacts |
| `fmt` | Format Go and Terraform code |
| `vet` | Run `go vet` |
| `lint` | Run `golangci-lint` |
| `test` | Run unit tests |
| `testacc` | Run acceptance tests |
| `docs` | Generate provider documentation |
| `tidy` | Tidy Go modules |
| `dev-setup` | Create `~/.terraformrc` with dev overrides |

## License

MIT
