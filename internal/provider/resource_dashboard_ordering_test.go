package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/pborman/uuid"
)

func TestAccResourceDashboardOrdering(t *testing.T) {
	uniqueID := uuid.NewRandom().String()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Test
			{
				Config: providerConfig +
					`
resource "squaredup_workspace" "application_workspace" {
  display_name = "Application Team - ` + uniqueID + `"
  description  = "Workspace with Dashboards for Application Team"
}

resource "squaredup_dashboard" "application1" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 1"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard" "application1_api" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 1 (API)"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard" "application2" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 2"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard" "application3" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 3"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard_ordering" "application_workspace" {
  workspace_id = squaredup_workspace.application_workspace.id
  order = jsonencode([
    {
      name = "Application 1"
      id   = "64833627-3283-499e-8f3a-4beb88edfd79"
      dashboardIdOrder = [
        {
          name = "API"
          id   = "72ff3be7-b83e-4b7e-82f2-df89832a463c"
          dashboardIdOrder = [
            squaredup_dashboard.application1_api.id
          ]
        },
        squaredup_dashboard.application1.id
      ]
    },
    squaredup_dashboard.application2.id,
    squaredup_dashboard.application3.id
  ])
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("squaredup_dashboard_ordering.application_workspace", "workspace_id"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard_ordering.application_workspace", "order"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_dashboard_ordering.application_workspace",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update Test
			{
				Config: providerConfig +
					`
resource "squaredup_workspace" "application_workspace" {
  display_name = "Application Team - ` + uniqueID + `"
  description  = "Workspace with Dashboards for Application Team"
}

resource "squaredup_dashboard" "application1" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 1"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard" "application1_api" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 1 (API)"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard" "application2" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 2"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard" "application3" {
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Application 3"
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "07bb1be5-e210-4fa1-81e7-728a750ed247",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "",
        "description": ""
      }
    }
  ]
}
EOT
}

resource "squaredup_dashboard_ordering" "application_workspace" {
  workspace_id = squaredup_workspace.application_workspace.id
  order = jsonencode([
    squaredup_dashboard.application3.id,
	squaredup_dashboard.application2.id,
	squaredup_dashboard.application1.id,
    squaredup_dashboard.application1_api.id
  ])
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("squaredup_dashboard_ordering.application_workspace", "workspace_id"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard_ordering.application_workspace", "order"),
				),
			},
		},
	})
}
