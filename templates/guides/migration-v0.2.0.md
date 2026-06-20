---
layout: "discord"
page_title: "Migrating to terraform-provider-discord v0.2.0"
description: |-
  This guide explains the changes in v0.2.0 and how to migrate from v0.1.0.
---

# Migration Guide: v0.1.0 to v0.2.0

This guide helps you migrate from terraform-provider-discord v0.1.0 to v0.2.0.

## New Resources

v0.2.0 introduces three new resources:

- `discord_emoji` - Manage custom emoji
- `discord_webhook` - Manage webhooks
- `discord_invite` - Manage server invites

## New Data Sources

v0.2.0 introduces one new data source:

- `discord_server` - Get information about a Discord server

## Resource Changes

### `discord_role` - Permissions Format Change

The `permissions` attribute now accepts permission bitwise integer values instead of permission names.

**Before (v0.1.0):**
```hcl
resource "discord_role" "example" {
  server_id   = discord_server.example.id
  name        = "Example Role"
  permissions = "view_channel,read_messages"
}
```

**After (v0.2.0):**
```hcl
resource "discord_role" "example" {
  server_id   = discord_server.example.id
  name        = "Example Role"
  permissions = "1024" # bitwise value for view_channel
}
```

Use the Discord developer documentation to calculate permission bitwise values, or use the `terraform console` to compute them.

### `discord_channel` - Type Attribute

The `type` attribute is now required and must be set to `0` for text channels. Channel types:
- `0` - Text channel
- `2` - Voice channel
- `4` - Category (planned for v0.3.0)
- `13` - Forum channel (planned for v0.3.0)

## Import Support

All resources now support import via `terraform import`.

Example:
```bash
terraform import discord_server.example 123456789012345678
```

## Deprecations

No resources are deprecated in v0.2.0.

## Upgrading

1. Update your `terraform.lock.hcl`:
   ```bash
   terraform init -upgrade
   ```

2. Run `terraform plan` to see the proposed changes.

3. Address any errors related to the permission format change in `discord_role` resources.

## Known Issues

- Channel permission overwrites (`discord_channel_permission`) are not yet supported due to discordgo API limitations
- Forum channels (`discord_forum_channel`) planned for v0.3.0
- Category channels (`discord_category`) planned for v0.3.0
