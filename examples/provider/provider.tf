terraform {
  required_providers {
    discord = {
      source  = "echohello-dev/discord"
      version = "~> 0.1"
    }
  }
}

provider "discord" {
  # token = "your-bot-token-here"
  # Or use DISCORD_TOKEN environment variable
}
