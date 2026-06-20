resource "discord_channel" "general" {
  server_id = discord_server.community.id
  name      = "general"
  type      = 0 # GUILD_TEXT
  topic     = "General discussion"
  position  = 0
}
