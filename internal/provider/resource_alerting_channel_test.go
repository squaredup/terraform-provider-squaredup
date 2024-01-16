package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceAlertingChannel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig +
					`
data "squaredup_alerting_channel_types" "example" {
	display_name = "Slack API"
}

resource "squaredup_alerting_channel" "slack_api_alert_channel_test" {
	display_name    = "Slack Alert - Team DevOps"
	channel_type_id = data.squaredup_alerting_channel_types.example.alerting_channel_types[0].channel_id
	config = jsonencode({
		channel = "devops"
		token   = "some-token"
	})
	enabled = true
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_alerting_channel.slack_api_alert_channel_test", "display_name", "Slack Alert - Team DevOps"),
					resource.TestCheckResourceAttrSet("squaredup_alerting_channel.slack_api_alert_channel_test", "id"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_alerting_channel.slack_api_alert_channel_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "config"},
			},
			// Update Test
			{
				Config: providerConfig +
					`
data "squaredup_alerting_channel_types" "example" {
	display_name = "Slack API"
}

resource "squaredup_alerting_channel" "slack_api_alert_channel_test" {
	display_name    = "Slack Alert - DevOps Team"
	channel_type_id = data.squaredup_alerting_channel_types.example.alerting_channel_types[0].channel_id
	config = jsonencode({
		channel = "devops"
		token   = "some-token"
	})
	enabled = true
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_alerting_channel.slack_api_alert_channel_test", "display_name", "Slack Alert - DevOps Team"),
				),
			},
		},
	})
}
