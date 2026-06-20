resource "discord_webhook" "alerts" {
  server_id  = discord_server.community.id
  channel_id = discord_channel.general.id
  name       = "alerts"
}
