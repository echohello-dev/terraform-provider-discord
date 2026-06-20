resource "discord_invite" "general" {
  server_id  = discord_server.community.id
  channel_id = discord_channel.general.id
  max_uses   = 10
  unique     = true
}
