package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/pborman/uuid"
)

func TestAccSquaredUpDashboardImageResource(t *testing.T) {
	uuid := uuid.NewRandom().String()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "squaredup_workspace" "application_workspace" {
	display_name = "Application Team - ` + uuid + `"
	description  = "Workspace with Dashboards for Application Team"
}

locals {
	application1_image_tile_id = "075c1fad-3c18-4aec-9a5c-c63b4478580e"
}

resource "squaredup_dashboard" "application_dashboard" {
	display_name = "Application Dashboard - ` + uuid + `"
	workspace_id = squaredup_workspace.application_workspace.id
	dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "contents": [
    {
      "x": 0,
      "h": 2,
      "i": "${local.application1_image_tile_id}",
      "y": 0,
      "config": {
        "_type": "tile/image",
        "description": "",
        "title": "",
        "visualisation": {
          "config": {
            "title": "Some description",
            "uploaded": 123456789,
            "showHealthState": false
          }
        }
      },
      "w": 4
    }
  ],
  "version": 5,
  "columns": 4
}
EOT
}

resource "squaredup_dashboard_image" "image_resource" {
	tile_id               = local.application1_image_tile_id
	dashboard_id          = squaredup_dashboard.application_dashboard.id
	workspace_id          = squaredup_workspace.application_workspace.id
	image_base64_data_uri = "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAgAAAAIACAYAAAD0eNT6AAAA"
	image_file_name       = "circle.png"
}
				  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check Fixed Values
					resource.TestCheckResourceAttr("squaredup_dashboard_image.image_resource", "image_file_name", "circle.png"),
					resource.TestCheckResourceAttr("squaredup_dashboard_image.image_resource", "image_base64_data_uri", "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAgAAAAIACAYAAAD0eNT6AAAA"),
				),
			},
			// Import Test
			{
				ResourceName:      "squaredup_dashboard_image.image_resource",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					workspaceId := state.RootModule().Resources["squaredup_workspace.application_workspace"].Primary.ID
					dashboardId := state.RootModule().Resources["squaredup_dashboard.application_dashboard"].Primary.ID
					tileId := state.RootModule().Resources["squaredup_dashboard_image.image_resource"].Primary.ID
					return fmt.Sprintf("%s,%s,%s", workspaceId, dashboardId, tileId), nil
				},
			},
			// Update Test
			{
				Config: providerConfig + `
resource "squaredup_workspace" "application_workspace" {
	display_name = "Application Team - ` + uuid + `"
	description  = "Workspace with Dashboards for Application Team"
}

locals {
	application1_image_tile_id = "075c1fad-3c18-4aec-9a5c-c63b4478580e"
}

resource "squaredup_dashboard" "application_dashboard" {
	display_name = "Application Dashboard - ` + uuid + `"
	workspace_id = squaredup_workspace.application_workspace.id
	dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "contents": [
    {
      "x": 0,
      "h": 2,
      "i": "${local.application1_image_tile_id}",
      "y": 0,
      "config": {
        "_type": "tile/image",
        "description": "",
        "title": "",
        "visualisation": {
          "config": {
            "title": "Some description",
            "uploaded": 123456789,
            "showHealthState": false
          }
        }
      },
      "w": 4
    }
  ],
  "version": 5,
  "columns": 4
}
EOT
}

resource "squaredup_dashboard_image" "image_resource" {
	tile_id               = local.application1_image_tile_id
	dashboard_id          = squaredup_dashboard.application_dashboard.id
	workspace_id          = squaredup_workspace.application_workspace.id
	image_base64_data_uri = "data:image/png;base64, EmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSpG=="
	image_file_name       = "square.png"
}
				  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard_image.image_resource", "image_file_name", "square.png"),
					resource.TestCheckResourceAttr("squaredup_dashboard_image.image_resource", "image_base64_data_uri", "data:image/png;base64, EmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSpG=="),
				),
			},
		},
	})
}
