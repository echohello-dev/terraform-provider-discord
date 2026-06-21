data "discord_role" "moderator" {
  server_id = "123456789012345678"
  id        = "987654321098765432"
}

output "role_name" {
  value = data.discord_role.moderator.name
}

output "role_permissions" {
  value = data.discord_role.moderator.permissions
}