resource "squaredup_workspace" "application_workspace" {
  display_name            = "Application Team"
  allow_dashboard_sharing = false
  description             = "Workspace with Dashboards for Application Team"
}

locals {
  // unique id for the image tile which is generated using uuidgen in bash
  application1_image_tile_id = "1d5ec28e-d5d5-4bc9-bbe0-dd63d9871244"
}

resource "squaredup_dashboard" "application1_dashboard" {
  display_name       = "Application 1"
  workspace_id       = squaredup_workspace.application_workspace.id
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

resource "squaredup_dashboard_image" "application1_image" {
  tile_id               = local.application1_image_tile_id
  dashboard_id          = squaredup_dashboard.application1_dashboard.id
  workspace_id          = squaredup_workspace.application_workspace.id
  image_base64_data_uri = "data:image/png;base64, ${filebase64("path/to/your/image.png")}"
  image_file_name       = "application1.png"
}
