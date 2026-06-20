resource "discord_emoji" "custom" {
  server_id = discord_server.community.id
  name      = "custom_emoji"
  image_url = "https://example.com/emoji.png"
}
