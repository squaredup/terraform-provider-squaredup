terraform {
  required_providers {
    squaredup = {
      source = "registry.terraform.io/squaredup/squaredup"
    }
  }
}

provider "squaredup" {
  region  = "us/eu"
  api_key = "api-key"
}

data "squaredup_datasources" "azure_devops" {
  data_source_name = "Azure DevOps"
}

data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "ado_datasource" {
  display_name     = "Azure DevOps"
  data_source_name = data.squaredup_datasources.azure_devops.plugins[0].display_name
  config = jsonencode({
    org         = "org-name"
    accessToken = "access-token"
  })
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
  display_name        = "DevOps Team"
  description         = "Workspace with Dashboards for DevOps Team"
  type                = "application"
  tags                = ["terraform", "auto-created"]
  open_access_enabled = true
  workspaces_links    = [squaredup_workspace.application_workspace.id]
  datasources_links   = [squaredup_datasource.ado_datasource.id, squaredup_datasource.sample_data_source.id]
}

data "squaredup_data_streams" "azure_devops_dataStreams" {
  data_source_id = data.squaredup_datasources.azure_devops.plugins[0].id
}

locals {
  build_runs  = data.squaredup_data_streams.azure_devops_dataStreams.data_streams[index(data.squaredup_data_streams.azure_devops_dataStreams.data_streams.*.definition_name, "buildruns")]
  agent_usage = data.squaredup_data_streams.azure_devops_dataStreams.data_streams[index(data.squaredup_data_streams.azure_devops_dataStreams.data_streams.*.definition_name, "agentusage")]
}

data "squaredup_data_streams" "sample_data_logs_dataStreams" {
  data_source_id              = data.squaredup_datasources.sample_data.plugins[0].id
  data_stream_definition_name = "logs"
}

data "squaredup_data_streams" "sample_data_lambdaerrors_dataStreams" {
  data_source_id              = data.squaredup_datasources.sample_data.plugins[0].id
  data_stream_definition_name = "perf-lambda-errors"
}

resource "squaredup_dashboard" "devops_team_dashboard" {
  dashboard_template = file("dashboardTemplate.json")
  template_bindings = jsonencode({
    azure_devops_data_source_id = squaredup_datasource.ado_datasource.id
    sample_data_source_id       = squaredup_datasource.sample_data_source.id
    cloud_watch_logs_id         = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[0].id
    azure_devops_build_runs_id  = local.build_runs.id
    azure_devops_agent_usage_id = local.agent_usage.id
    perf_lambda_errors_id       = data.squaredup_data_streams.sample_data_lambdaerrors_dataStreams.data_streams[0].id
  })
  workspace_id = squaredup_workspace.devops_workspace.id
  display_name = "Build Stats and Logs"
}
