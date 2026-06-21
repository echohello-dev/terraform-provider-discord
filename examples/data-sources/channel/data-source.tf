data "discord_channel" "general" {
  id = "123456789012345678"
}

output "channel_name" {
  value = data.discord_channel.general.name
}

output "channel_topic" {
  value = data.discord_channel.general.topic
}