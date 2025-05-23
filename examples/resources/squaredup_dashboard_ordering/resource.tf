resource "squaredup_workspace" "application_workspace" {
  display_name = "Application Team"
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

resource "random_uuid" "application_1_folder_id" {}

resource "random_uuid" "app1_api_folder_id" {}

resource "squaredup_dashboard_ordering" "example_ordering" {
  workspace_id = squaredup_workspace.application_workspace.id
  # Example with dashboards in top level and nested folders
  # Folders needs to be as an object with name, id and dashboardIdOrder
  # dashboardIdOrder is a list of dashboard ids or nested folders
  order = jsonencode([
    {
      name = "Application 1"
      id   = random_uuid.application_1_folder_id.result
      dashboardIdOrder = [
        {
          name = "API"
          id   = random_uuid.app1_api_folder_id.result
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
