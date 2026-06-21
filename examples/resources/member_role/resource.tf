resource "discord_member_role" "mod_assignment" {
  server_id = discord_server.community.id
  user_id   = "111222333444555666"
  role_id   = discord_role.moderator.id
}