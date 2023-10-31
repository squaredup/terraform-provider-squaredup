data "squaredup_alerting_channel_types" "example" {
  display_name = "Slack API"
}

resource "squaredup_alerting_channel" "slack_api_alert" {
  display_name    = "Slack Alert - Team DevOps"
  channel_type_id = data.squaredup_alerting_channel_types.example.alerting_channel_types[0].channel_id
  config = jsonencode({
    channel = "devops"
    token   = "some-token"
  })
  enabled = true
}
