resource "discord_server" "community" {
  name                          = "My Terraform Community"
  region                        = "us-west"
  verification_level            = 1
  default_message_notifications = 0
  explicit_content_filter       = 2
  afk_timeout                   = 300
}
