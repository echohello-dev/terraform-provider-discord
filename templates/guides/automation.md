---
layout: "discord"
page_title: "Automation with terraform-provider-discord"
description: |-
  This guide explains how to automate Discord server management using terraform-provider-discord.
---

# Automation Guide

This guide covers automation patterns for managing Discord infrastructure with terraform-provider-discord.

## CI/CD Integration

### GitHub Actions

Example workflow to apply Terraform on push to main:

```yaml
name: Discord Infrastructure

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.9.0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Install provider
        run: go install .

      - name: Setup Terraform
        env:
          DISCORD_TOKEN: ${{ secrets.DISCORD_TOKEN }}
        run: |
          terraform init
          terraform validate

      - name: Terraform Plan
        env:
          DISCORD_TOKEN: ${{ secrets.DISCORD_TOKEN }}
        run: terraform plan -no-color

      - name: Terraform Apply
        if: github.ref == 'refs/heads/main' && github.event_name == 'push'
        env:
          DISCORD_TOKEN: ${{ secrets.DISCORD_TOKEN }}
        run: terraform apply -auto-approve
```

### Required Secrets

| Secret | Description |
|--------|-------------|
| `DISCORD_TOKEN` | Discord bot token with appropriate permissions |

## State Management

### Remote State

Use Terraform Cloud or S3 for state storage:

```hcl
terraform {
  backend "remote" {
    organization = "your-org"
    workspaces {
      name = "discord-infrastructure"
    }
  }
}
```

### State Locking

Remote backends provide state locking automatically. For local state, use:
```bash
terraform apply -lock=true
```

## Environment-Specific Configurations

### Using Workspaces

Create separate workspaces for dev/staging/prod:

```bash
terraform workspace new dev
terraform workspace new staging
terraform workspace new prod
```

Reference the workspace in your configuration:

```hcl
locals {
  environment = terraform.workspace
}

resource "discord_role" "admin" {
  server_id = var.server_ids[local.environment]
  name      = "${local.environment}-admin"
  # ...
}
```

## Role-Based Access Patterns

### Hierarchical Roles

```hcl
resource "discord_role" "admin" {
  server_id   = discord_server.main.id
  name        = "Admin"
  permissions = "1073741823" # all permissions
  hoist       = true
  color       = 15158332 # red
}

resource "discord_role" "moderator" {
  server_id   = discord_server.main.id
  name        = "Moderator"
  permissions = "66321471"
  hoist       = true
  color       = 3066993 # green
}

resource "discord_role" "member" {
  server_id   = discord_server.main.id
  name        = "Member"
  permissions = "104320001"
}
```

### Role Assignment via Groups

Use Terraform to manage role assignments through external systems or Discord's API directly.

## Webhook Integrations

### Notification Webhook

```hcl
resource "discord_webhook" "notifications" {
  channel_id = discord_channel.notifications.id
  name      = "CI Notifications"
}

output "webhook_url" {
  value     = "https://discord.com/api/webhooks/${discord_webhook.notifications.id}/${discord_webhook.notifications.token}"
  sensitive = true
}
```

### Multi-Server Notifications

```hcl
locals {
  servers = {
    "guild-1" = { id = "123456789", webhook_channel = "channel-id-1" }
    "guild-2" = { id = "987654321", webhook_channel = "channel-id-2" }
  }
}

resource "discord_webhook" "notifications" {
  for_each   = local.servers
  channel_id = each.value.webhook_channel
  name       = "Notifications - ${each.key}"
}
```

## Importing Existing Resources

Import existing Discord resources into Terraform state:

```bash
# Import a server
terraform import discord_server.main 123456789012345678

# Import a channel
terraform import discord_channel.announcements 123456789_987654321

# Import a role
terraform import discord_role.admin 123456789_111222333444555

# Import an emoji
terraform import discord_emoji.custom_emoji 123456789_666777888
```

Format for import IDs:
- Server: `{guild_id}`
- Channel: `{guild_id}_{channel_id}`
- Role: `{guild_id}_{role_id}`
- Emoji: `{guild_id}_{emoji_id}`
- Webhook: `{webhook_id}`
- Invite: `{invite_code}`

## Testing

### Acceptance Tests

Run acceptance tests against a test bot:

```bash
export TF_ACC=1
export DISCORD_TOKEN=your-test-bot-token
go test -v ./...
```

### Plan Validation in CI

Always run `terraform plan` in CI before applying to catch configuration errors:

```bash
terraform plan -out=tfplan
# In PR review, show tfplan output
terraform show -json tfplan | jq '.changes'
```

## Best Practices

1. **Use remote state** - Never store Terraform state in version control
2. **Lock your provider version** - Pin to a specific version in your configuration
3. **Use workspaces** - Separate environments (dev/staging/prod)
4. **Import existing resources** - Bring existing Discord resources under Terraform management
5. **Protect sensitive values** - Use `sensitive = true` for webhook tokens and bot tokens
6. **Use ` RequiresReplace`** - Plan modifier on fields that require resource recreation (e.g., server_id)
