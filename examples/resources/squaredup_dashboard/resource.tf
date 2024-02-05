data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
  display_name      = "Application Team"
  description       = "Workspace with Dashboards for Application Team"
  datasources_links = [squaredup_datasource.sample_data_source.id]
}

data "squaredup_data_streams" "sample_data_logs_dataStreams" {
  data_source_id = data.squaredup_datasources.sample_data.plugins[0].id
}

locals {
  logs_data_stream               = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "logs")]
  perf_lambda_errors_data_stream = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "perf-lambda-errors")]
  cost_data_stream               = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "perf-cost")]
}

data "squaredup_nodes" "acommon_node" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_name      = "account-common-lambda"
}

resource "squaredup_dashboard" "sample_dashboard" {
  dashboard_template = <<EOT
{
  "_type": "layout/grid",
  "contents": [
    {
      "static": false,
      "w": 2,
      "moved": false,
      "h": 3,
      "x": 0,
      "y": 0,
      "i": "1",
      "z": 0,
      "config": {
        "dataStream": {
          "pluginConfigId": "{{sample_data_source_id}}",
          "id": "{{cloud_watch_logs_id}}"
        },
        "scope": {
          "query": "g.V().order().by('__name').hasNot('__canonicalType').has(\"__configId\", \"{{sample_data_source_id}}\").or(__.has(\"sourceType\", within(\"sample-function\",\"sample-server\",\"sample-database\"))).limit(500)",
          "bindings": {},
          "queryDetail": {}
        },
        "_type": "tile/data-stream",
        "description": "",
        "baseTile": "data-stream-base-tile",
        "title": "CloudWatch Logs",
        "visualisation": {
          "type": "data-stream-table",
          "config": {
            "data-stream-table": {
              "resizedColumns": {
                "columnWidths": {
                  "logs.timestamp": 146
                }
              }
            }
          }
        }
      }
    },
    {
      "static": false,
      "w": 2,
      "moved": false,
      "h": 3,
      "x": 2,
      "y": 0,
      "i": "a8255dce-5f74-4ff5-b3d3-138f6a0ff130",
      "z": 0,
      "config": {
        "_type": "tile/data-stream",
        "description": "",
        "title": "Lambda Errors",
        "dataStream": {
          "pluginConfigId": "{{sample_data_source_id}}",
          "filter": {
            "multiOperation": "and",
            "filters": []
          },
          "id": "{{perf_lambda_errors_id}}",
          "group": {
            "by": [
              "data.lambdaErrors.label",
              "uniqueValues"
            ],
            "aggregate": [
              {
                "type": "sum",
                "names": [
                  "data.lambdaErrors.value"
                ]
              }
            ]
          }
        },
        "visualisation": {
          "type": "data-stream-donut-chart"
        },
        "scope": {
          "query": "g.V().order().by('__name').hasNot('__canonicalType').has(\"__configId\", \"{{sample_data_source_id}}\").or(__.has(\"sourceType\", \"sample-function\")).limit(500)",
          "bindings": {},
          "queryDetail": {}
        }
      }
    },
    {
      "static": false,
      "w": 2,
      "moved": false,
      "h": 3,
      "x": 0,
      "y": 3,
      "i": "aec96894-63e6-4873-89f8-22df1c10d5d0",
      "z": 0,
      "config": {
        "title": "Account Common Lambda Cost",
        "_type": "tile/data-stream",
        "monitor": {
          "_type": "simple",
          "tileRollsUp": true,
          "monitorType": "threshold",
          "frequency": 720,
          "aggregation": "top",
          "column": "data.cost.value_sum",
          "condition": {
            "columns": [
              "data.cost.value_sum"
            ],
            "logic": {
              "if": [
                {
                  ">": [
                    {
                      "var": "top"
                    },
                    500
                  ]
                },
                "error",
                {
                  ">": [
                    {
                      "var": "top"
                    },
                    400
                  ]
                },
                "warning"
              ]
            }
          }
        },
        "dataStream": {
          "pluginConfigId": "{{sample_data_source_id}}",
          "id": "{{cost_data_stream}}",
          "group": {
            "by": [
              "data.cost.label",
              "uniqueValues"
            ],
            "aggregate": [
              {
                "type": "sum",
                "names": [
                  "data.cost.value"
                ]
              }
            ]
          }
        },
        "visualisation": {
          "type": "data-stream-scalar"
        },
        "scope": {
          "query": "g.V().has('id', within(ids_xAvxTqo9n9QCEeCHq2d1)).has(\"__configId\", \"{{sample_data_source_id}}\").or(__.has(\"sourceType\", within(\"sample-function\",\"sample-server\")))",
          "bindings": {
            "ids_xAvxTqo9n9QCEeCHq2d1": [
              "{{acommon_node_id}}"
            ]
          },
          "queryDetail": {
            "ids": [
              "{{acommon_node_id}}"
            ]
          }
        }
      }
    }
  ],
  "version": 1,
  "columns": 4
}
EOT
  template_bindings = jsonencode({
    sample_data_source_id = squaredup_datasource.sample_data_source.id
    cloud_watch_logs_id   = local.logs_data_stream.id
    perf_lambda_errors_id = local.perf_lambda_errors_data_stream.id
    cost_data_stream      = local.cost_data_stream.id
    acommon_node_id       = data.squaredup_nodes.acommon_node.node_properties[0].id
  })
  workspace_id = squaredup_workspace.application_workspace.id
  display_name = "Sample Dashboard"
  timeframe    = "last12hours"
}
