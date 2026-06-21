resource "discord_message" "welcome" {
  channel_id = discord_channel.general.id
  content    = "Welcome to the server!"
}