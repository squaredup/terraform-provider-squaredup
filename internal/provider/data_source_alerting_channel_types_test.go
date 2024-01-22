package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceAlertingChannelTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig +
					`
data "squaredup_alerting_channel_types" "alert_channel_slack" {
	display_name = "Slack API"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.squaredup_alerting_channel_types.alert_channel_slack", "display_name", "Slack API"),
					resource.TestCheckResourceAttrSet("data.squaredup_alerting_channel_types.alert_channel_slack", "alerting_channel_types.0.channel_id"),
				),
			},
		},
	})
}
