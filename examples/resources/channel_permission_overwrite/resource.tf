resource "discord_channel_permission_overwrite" "private_general" {
  channel_id  = discord_channel.general.id
  target_id   = discord_role.moderator.id
  target_type = "role"
  allow       = "1024" # VIEW_CHANNEL
  deny        = "2048" # SEND_MESSAGES
}