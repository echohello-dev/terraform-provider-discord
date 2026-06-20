data "discord_server" "existing" {
  id = "123456789012345678"
}

output "server_name" {
  value = data.discord_server.existing.name
}
