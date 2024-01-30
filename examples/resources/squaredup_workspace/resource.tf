data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
  display_name = "Application Team"
  description  = "Workspace with Dashboards for Application Team"
}

resource "squaredup_workspace" "devops_workspace" {
  display_name            = "DevOps Team"
  description             = "Workspace with Dashboards for DevOps Team"
  type                    = "application"
  tags                    = ["terraform", "auto-created"]
  allow_dashboard_sharing = true
  workspaces_links        = [squaredup_workspace.application_workspace.id]
  datasources_links       = [squaredup_datasource.sample_data_source.id]
}
