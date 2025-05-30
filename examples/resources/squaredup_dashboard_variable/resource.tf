data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

data "squaredup_data_streams" "sample_data_logs_dataStreams" {
  data_source_id = data.squaredup_datasources.sample_data.plugins[0].id
}

resource "squaredup_workspace" "application_workspace" {
  display_name      = "Application Team"
  description       = "Workspace with Dashboards for Application Team"
  datasources_links = [squaredup_datasource.sample_data_source.id]
  tags              = ["dashboard-variable"]
}

resource "squaredup_scope" "dynamic_scope" {
  scope_type     = "dynamic"
  display_name   = "Dynamic Scope"
  workspace_id   = squaredup_workspace.application_workspace.id
  data_source_id = [squaredup_datasource.sample_data_source.id]
  types          = ["sample-function"]
  search_query   = "account-common"
}

resource "squaredup_dashboard_variable" "example_all_variable" {
  workspace_id             = squaredup_workspace.application_workspace.id
  collection_id            = squaredup_scope.dynamic_scope.id
  default_object_selection = "all" # or "none"
}

resource "squaredup_dashboard" "all_objects" {
  workspace_id          = squaredup_workspace.application_workspace.id
  display_name          = "All Objects"
  dashboard_variable_id = squaredup_dashboard_variable.example_all_variable.id
  dashboard_template    = <<EOT
{
  "_type": "layout/grid",
  "contents": [
    {
      "x": 0,
      "h": 2,
      "i": "ebd9b9cd-2fb3-4978-9a7e-0c96a70a6e41",
      "y": 0,
      "config": {
        "variables": [
          "{{example_variable_id}}"
        ],
        "scope": {
          "workspace": "{{workspace_id}}",
          "scope": "{{dynamic_scope_id}}"
        },
        "dataStream": {
          "name": "perf-lambda-duration",
          "id": "{{perf_lambda_duration_datastream_id}}"
        },
        "_type": "tile/data-stream",
        "description": "",
        "title": "Lambda Duration",
        "visualisation": {
          "type": "data-stream-line-graph",
          "config": {
            "data-stream-line-graph": {
              "seriesColumn": "data.lambdaDuration.label",
              "xAxisColumn": "data.timestamp",
              "yAxisColumn": "data.lambdaDuration.value"
            }
          }
        }
      },
      "w": 4
    }
  ],
  "version": 1,
  "columns": 4
}
EOT
  template_bindings = jsonencode({
    example_variable_id                = squaredup_dashboard_variable.example_variable.id
    workspace_id                       = squaredup_workspace.application_workspace.id
    dynamic_scope_id                   = squaredup_scope.dynamic_scope.id
    perf_lambda_duration_datastream_id = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "perf-lambda-duration")].id
  })
}

resource "squaredup_dashboard_variable" "example_none_variable" {
  workspace_id                    = squaredup_workspace.application_workspace.id
  collection_id                   = squaredup_scope.dynamic_scope.id
  default_object_selection        = "none"
  allow_multiple_object_selection = true
}

resource "squaredup_dashboard" "no_default_objects" {
  workspace_id          = squaredup_workspace.application_workspace.id
  display_name          = "No Default Objects"
  dashboard_variable_id = squaredup_dashboard_variable.example_none_variable.id
  dashboard_template    = <<EOT
{
  "_type": "layout/grid",
  "columns": 4,
  "contents": [
    {
      "i": "5cbfb630-ba91-445a-aaaa-eb7fe9161ca0",
      "x": 0,
      "y": 0,
      "w": 4,
      "h": 2,
      "config": {
        "title": "Cost",
        "description": "",
        "_type": "tile/data-stream",
        "variables": [
          "{{example_none_variable_id}}"
        ],
        "scope": {
          "scope": "{{dynamic_scope_id}}",
          "workspace": "{{workspace_id}}"
        },
        "dataStream": {
          "id": "{{perf_cost_datastream_id}}",
          "name": "perf-cost"
        },
        "visualisation": {
          "type": "data-stream-line-graph",
          "config": {
            "data-stream-line-graph": {
              "xAxisColumn": "data.timestamp",
              "yAxisColumn": "data.cost.value",
              "seriesColumn": "data.cost.label"
            }
          }
        }
      }
    }
  ],
  "version": 1
}
EOT
  template_bindings = jsonencode({
    example_none_variable_id = squaredup_dashboard_variable.example_none_variable.id
    workspace_id             = squaredup_workspace.application_workspace.id
    dynamic_scope_id         = squaredup_scope.dynamic_scope.id
    perf_cost_datastream_id  = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "perf-cost")].id
  })
}
