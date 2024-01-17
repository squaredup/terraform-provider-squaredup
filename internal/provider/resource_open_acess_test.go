package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceOpenAccess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig +
					`
resource "squaredup_workspace" "application_workspace" {
	display_name        = "OA Test Workspace"
	description         = "Workspace with Dashboards for Application Team"
	open_access_enabled = true
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
			"content": "Sample Tile",
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
	workspace_id = squaredup_workspace.application_workspace.id
	display_name = "Sample Dashboard - OA Test"
}

resource "squaredup_dashboard_share" "sample_dashboard_share" {
	dashboard_id           = squaredup_dashboard.sample_dashboard.id
	workspace_id           = squaredup_workspace.application_workspace.id
	require_authentication = true
	enable_link            = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard_share.sample_dashboard_share", "require_authentication", "true"),
					resource.TestCheckResourceAttr("squaredup_dashboard_share.sample_dashboard_share", "enable_link", "true"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard_share.sample_dashboard_share", "id"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard_share.sample_dashboard_share", "open_access_link"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_dashboard_share.sample_dashboard_share",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update Test
			{
				Config: providerConfig +
					`
resource "squaredup_workspace" "application_workspace" {
display_name        = "OA Test Workspace"
description         = "Workspace with Dashboards for Application Team"
open_access_enabled = true
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
		"content": "Sample Tile",
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
workspace_id = squaredup_workspace.application_workspace.id
display_name = "Sample Dashboard - OA Test"
}

resource "squaredup_dashboard_share" "sample_dashboard_share" {
dashboard_id           = squaredup_dashboard.sample_dashboard.id
workspace_id           = squaredup_workspace.application_workspace.id
require_authentication = false
enable_link            = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard_share.sample_dashboard_share", "require_authentication", "false"),
					resource.TestCheckResourceAttr("squaredup_dashboard_share.sample_dashboard_share", "enable_link", "false"),
				),
			},
		},
	})
}
