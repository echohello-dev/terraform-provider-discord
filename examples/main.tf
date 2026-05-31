terraform {
  required_providers {
    discord = {
      source  = "echohello-dev/discord"
      version = "0.1.0"
    }
  }
}

provider "discord" {
  # token = "your-bot-token-here"
  # Or use DISCORD_TOKEN environment variable
}

# Data source to read existing server
# data "discord_server" "existing" {
#   id = "123456789012345678"
# }

resource "discord_server" "community" {
  name                          = "My Terraform Community"
  region                        = "us-west"
  verification_level            = 1
  default_message_notifications = 0
  explicit_content_filter       = 2
  afk_timeout                   = 300
}

resource "discord_channel" "general" {
  server_id = discord_server.community.id
  name      = "general"
  type      = 0 # GUILD_TEXT
  topic     = "General discussion"
  position  = 0
}

resource "discord_channel" "voice" {
  server_id  = discord_server.community.id
  name       = "General Voice"
  type       = 2 # GUILD_VOICE
  bitrate    = 64000
  user_limit = 0
}

resource "discord_role" "moderator" {
  server_id   = discord_server.community.id
  name        = "Moderator"
  permissions = "268435456" # Manage Messages etc.
  color       = 3447003     # Blurple
  hoist       = true
  mentionable = true
}
