package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/pborman/uuid"
)

func TestDashboardResource(t *testing.T) {
	uuid := uuid.NewRandom().String()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			//Create Dashboard Test
			{
				Config: providerConfig + `
resource "squaredup_workspace" "application_workspace" {
	display_name        = "Dashboard Test - ` + uuid + `"
	description         = "Workspace with Dashboards for Application Team"
}

resource "squaredup_dashboard" "sample_dashboard" {
	dashboard_template = <<EOT
{
"_type": "layout/grid",
"contents": [
	{
	"x": 0,
	"h": 2,
	"i": "1",
	"y": 0,
	"config": {
		"title": "",
		"description": "",
		"_type": "tile/text",
		"visualisation": {
		"config": {
			"content": "{{tile_text}}",
			"autoSize": true,
			"fontSize": 16,
			"align": "center"
		}
		}
	},
	"w": 4
	}
],
"columns": 1,
"version": 1
}
EOT
	template_bindings = jsonencode({
		tile_text = "Hello World"
	})
	workspace_id = squaredup_workspace.application_workspace.id
	timeframe = "last12hours"
	display_name = "Sample Dashboard - Dashboard Test - ` + uuid + `"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard.sample_dashboard", "display_name", "Sample Dashboard - Dashboard Test - "+uuid),
					resource.TestCheckResourceAttr("squaredup_dashboard.sample_dashboard", "timeframe", "last12hours"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard.sample_dashboard", "dashboard_content"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard.sample_dashboard", "schema_version"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard.sample_dashboard", "id"),
				),
			},
			//Import Dashboard Test
			{
				ResourceName:            "squaredup_dashboard.sample_dashboard",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "dashboard_content", "dashboard_template", "template_bindings"},
			},
			//Update Dashboard Test
			{
				Config: providerConfig + `
resource "squaredup_workspace" "application_workspace" {
	display_name        = "Dashboard Test - ` + uuid + `"
	description         = "Workspace with Dashboards for Application Team"
}

resource "squaredup_dashboard" "sample_dashboard" {
	dashboard_template = <<EOT
{
"_type": "layout/grid",
"contents": [
	{
	"x": 0,
	"h": 2,
	"i": "1",
	"y": 0,
	"config": {
		"title": "",
		"description": "",
		"_type": "tile/text",
		"visualisation": {
		"config": {
			"content": "{{tile_text}}",
			"autoSize": true,
			"fontSize": 16,
			"align": "center"
		}
		}
	},
	"w": 4
	}
],
"columns": 1,
"version": 1
}
EOT
	template_bindings = jsonencode({
		tile_text = "Hello World"
	})
	workspace_id = squaredup_workspace.application_workspace.id
	timeframe = "last1hour"
	display_name = "Sample Dashboard - Dashboard Test - ` + uuid + `Updated"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard.sample_dashboard", "display_name", "Sample Dashboard - Dashboard Test - "+uuid+"Updated"),
					resource.TestCheckResourceAttr("squaredup_dashboard.sample_dashboard", "timeframe", "last1hour"),
				),
			},
		},
	})
}
