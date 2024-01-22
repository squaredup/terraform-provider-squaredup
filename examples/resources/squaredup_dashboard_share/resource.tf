resource "squaredup_workspace" "application_workspace" {
  display_name        = "Application Team"
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
  workspace_id       = squaredup_workspace.application_workspace.id
  display_name       = "Sample Dashboard"
}

resource "squaredup_dashboard_share" "sample_dashboard_share" {
  dashboard_id           = squaredup_dashboard.sample_dashboard.id
  workspace_id           = squaredup_workspace.application_workspace.id
  require_authentication = true
  enabled                = true
}
