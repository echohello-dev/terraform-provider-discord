# Terraform Provider for Discord

A Terraform provider for managing Discord resources such as servers (guilds), channels, roles, and more using infrastructure-as-code principles.

## Features

- **Resources**
  - `discord_server` - Manage a Discord server (guild)
  - `discord_channel` - Manage a channel within a server
  - `discord_role` - Manage a role within a server
  - `discord_emoji` - Manage custom emojis in a server
  - `discord_webhook` - Manage webhooks for channels
  - `discord_invite` - Manage channel invites

- **Data Sources**
  - `discord_server` - Read information about an existing server

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.9
- Go >= 1.25 (for building from source)
- A Discord bot token with appropriate permissions

## Installation

### Terraform Registry

The provider is available on the Terraform Registry. Add to your Terraform configuration:

```hcl
terraform {
  required_providers {
    discord = {
      source  = "echohello-dev/discord"
      version = "~> 0.1"
    }
  }
}
```

### Pre-built Binaries

Download the latest release for your platform from the [GitHub Releases](https://github.com/echohello-dev/terraform-provider-discord/releases) page.

### From Source

```bash
git clone https://github.com/echohello-dev/terraform-provider-discord.git
cd terraform-provider-discord
go build -o terraform-provider-discord
```

## Configuration

### Provider Configuration Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `token` | string | Yes | - | Discord bot token |

### Environment Variables

```bash
export DISCORD_TOKEN="your-bot-token"
```

## Usage Examples

### Manage a Discord Server

```hcl
resource "discord_server" "community" {
  name                          = "My Terraform Community"
  region                        = "us-west"
  verification_level            = 1
  default_message_notifications = 0
  explicit_content_filter       = 2
  afk_timeout                   = 300
}
```

### Create Channels

```hcl
resource "discord_channel" "general" {
  server_id = discord_server.community.id
  name      = "general"
  type      = 0 # GUILD_TEXT
  topic     = "General discussion"
}

resource "discord_channel" "voice" {
  server_id = discord_server.community.id
  name      = "General Voice"
  type      = 2 # GUILD_VOICE
}
```

### Create Roles

```hcl
resource "discord_role" "moderator" {
  server_id   = discord_server.community.id
  name        = "Moderator"
  permissions = "268435456"
  color       = 3447003
  hoist       = true
  mentionable = true
}
```

### Read an Existing Server

```hcl
data "discord_server" "existing" {
  id = "123456789012345678"
}

output "server_name" {
  value = data.discord_server.existing.name
}
```

See the `examples/` directory for more complete usage examples.

## Development

### Build the provider

```bash
mise run build
```

### Install the provider locally

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
mise run doc
```

Or via `go generate`:

```bash
go generate ./...
```

### Tidy modules

```bash
mise run tidy
```

## Available Tasks

Run `mise run` (or `mise run h`) to see all available tasks.

| Task | Description |
|------|-------------|
| `build` | Build the provider binary to `.build/` |
| `install` | Install the provider to the local Terraform plugin directory |
| `clean` | Remove built artifacts |
| `fmt` | Format Go and Terraform code |
| `vet` | Run `go vet` |
| `lint` | Run `golangci-lint` |
| `test` | Run unit tests |
| `testacc` | Run acceptance tests |
| `doc` | Generate provider documentation |
| `validate` | Run lint + test |
| `tidy` | Tidy Go modules |
| `dev-setup` | Create `~/.terraformrc` with dev overrides |

## Debugging

Enable debug logging:

```bash
TF_LOG=DEBUG terraform plan
```

## Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

## License

MIT License - see LICENSE file for details.
