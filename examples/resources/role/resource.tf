resource "discord_role" "moderator" {
  server_id   = discord_server.community.id
  name        = "Moderator"
  permissions = "268435456"
  color       = 3447003
  hoist       = true
  mentionable = true
}
