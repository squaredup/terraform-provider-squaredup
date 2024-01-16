package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceWorkSpaceAlert(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig +
					`
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data Workspace Alert Test"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
	display_name      = "Application Team Workspace Alert Test"
	description       = "Workspace with Dashboards for Application Team"
	datasources_links = [squaredup_datasource.sample_data_source.id]
}

data "squaredup_data_streams" "sample_data_logs_dataStreams" {
	data_source_id = data.squaredup_datasources.sample_data.plugins[0].id
}

locals {
	logs_data_stream               = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "logs")]
	perf_lambda_errors_data_stream = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "perf-lambda-errors")]
}

resource "squaredup_dashboard" "sample_dashboard" {
	dashboard_template = <<EOT
{
	"_type": "layout/grid",
	"contents": [
		{
			"w": 2,
			"h": 3,
			"x": 0,
			"y": 0,
			"i": "1",
			"moved": false,
			"static": false,
			"config": {
				"baseTile": "data-stream-base-tile",
				"visualisation": {
					"config": {
						"data-stream-table": {
							"resizedColumns": {
								"columnWidths": {
									"logs.timestamp": 146
								}
							}
						}
					},
					"type": "data-stream-table"
				},
				"title": "CloudWatch Logs",
				"description": "",
				"_type": "tile/data-stream",
				"dataStream": {
					"id": "{{cloud_watch_logs_id}}",
					"pluginConfigId": "{{sample_data_source_id}}"
				},
				"scope": {
					"query": "g.V().order().by('__name').hasNot('__canonicalType').has(\"__configId\", \"{{sample_data_source_id}}\").or(__.has(\"sourceType\", within(\"sample-function\",\"sample-server\",\"sample-database\"))).limit(500)",
					"bindings": {},
					"queryDetail": {}
				}
			}
		},
		{
			"w": 2,
			"h": 3,
			"x": 2,
			"y": 0,
			"i": "a8255dce-5f74-4ff5-b3d3-138f6a0ff130",
			"moved": false,
			"static": false,
			"config": {
				"title": "Lambda Errors",
				"description": "",
				"_type": "tile/data-stream",
				"dataStream": {
					"id": "{{perf_lambda_errors_id}}",
					"pluginConfigId": "{{sample_data_source_id}}",
					"group": {
						"by": [
							"data.lambdaErrors.label",
							"uniqueValues"
						],
						"aggregate": [
							{
								"names": [
									"data.lambdaErrors.value"
								],
								"type": "sum"
							}
						]
					},
					"filter": {
						"filters": [],
						"multiOperation": "and"
					}
				},
				"visualisation": {
					"type": "data-stream-donut-chart"
				},
				"scope": {
					"query": "g.V().order().by('__name').hasNot('__canonicalType').has(\"__configId\", \"{{sample_data_source_id}}\").or(__.has(\"sourceType\", \"sample-function\")).limit(500)",
					"bindings": {},
					"queryDetail": {}
				},
				"monitor": {
					"_type": "simple",
					"tileRollsUp": true,
					"monitorType": "threshold",
					"frequency": 15,
					"aggregation": "top",
					"column": "data.lambdaErrors.value_sum",
					"condition": {
						"columns": [
							"data.lambdaErrors.value_sum"
						],
						"logic": {
							"if": [
								{
									">": [
										{
											"var": "top"
										},
										0
									]
								},
								"error"
							]
						}
					}
				}
			}
		}
	],
	"columns": 4,
	"version": 1
}
EOT
	template_bindings = jsonencode({
	sample_data_source_id = squaredup_datasource.sample_data_source.id
	cloud_watch_logs_id   = local.logs_data_stream.id
	perf_lambda_errors_id = local.perf_lambda_errors_data_stream.id
	})
	workspace_id = squaredup_workspace.application_workspace.id
	display_name = "Sample Dashboard for Workspace Alert Test"
	timeframe    = "last12hours"
}

# Extract ids of tiles
locals {
	dashboard_content = jsondecode(squaredup_dashboard.sample_dashboard.dashboard_content)

	lambda_errors_tile    = [for content in local.dashboard_content.contents : content.i if content.config.title == "Lambda Errors"]
	lambda_errors_tile_id = length(local.lambda_errors_tile) > 0 ? local.lambda_errors_tile[0] : null
}

resource "squaredup_alerting_channel" "slack_api_alert" {
	display_name    = "Slack Alert - Team DevOps for Workspace Alert Test"
	channel_type_id = "channeltype-00000000000000000001"
	config = jsonencode({
	channel = "devops"
	token   = "some-token"
	})
	enabled = true
}

resource "squaredup_workspace_alert" "example" {
	workspace_id = squaredup_workspace.application_workspace.id
	alerting_rules = [
	{
		channel   = squaredup_alerting_channel.slack_api_alert.id
		notify_on = "workspace_state"
		// "workspace_state" does not support "preview_image"
	},
	{
		channel       = squaredup_alerting_channel.slack_api_alert.id
		preview_image = true
		notify_on     = "all_monitors"
	},
	{
		channel       = squaredup_alerting_channel.slack_api_alert.id
		preview_image = false
		notify_on     = "selected_monitors"
		selected_monitors = [
		{
			dashboard_id = squaredup_dashboard.sample_dashboard.id
			tiles_id     = [local.lambda_errors_tile_id]
		}
		]
	}
	]
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace_alert.example", "alerting_rules.#", "3"),
					resource.TestCheckResourceAttrSet("squaredup_workspace_alert.example", "alerting_rules.0.channel"),
					resource.TestCheckResourceAttrSet("squaredup_workspace_alert.example", "workspace_id"),
					resource.TestCheckResourceAttr("squaredup_workspace_alert.example", "alerting_rules.0.notify_on", "workspace_state"),
					resource.TestCheckResourceAttr("squaredup_workspace_alert.example", "alerting_rules.0.preview_image", "false"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_workspace_alert.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update Test
			{
				Config: providerConfig +
					`
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data Workspace Alert Test"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
	display_name      = "Application Team Workspace Alert Test"
	description       = "Workspace with Dashboards for Application Team"
	datasources_links = [squaredup_datasource.sample_data_source.id]
}

data "squaredup_data_streams" "sample_data_logs_dataStreams" {
	data_source_id = data.squaredup_datasources.sample_data.plugins[0].id
}

locals {
	logs_data_stream               = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "logs")]
	perf_lambda_errors_data_stream = data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams[index(data.squaredup_data_streams.sample_data_logs_dataStreams.data_streams.*.definition_name, "perf-lambda-errors")]
}

resource "squaredup_dashboard" "sample_dashboard" {
	dashboard_template = <<EOT
{
	"_type": "layout/grid",
	"contents": [
		{
			"w": 2,
			"h": 3,
			"x": 0,
			"y": 0,
			"i": "1",
			"moved": false,
			"static": false,
			"config": {
				"baseTile": "data-stream-base-tile",
				"visualisation": {
					"config": {
						"data-stream-table": {
							"resizedColumns": {
								"columnWidths": {
									"logs.timestamp": 146
								}
							}
						}
					},
					"type": "data-stream-table"
				},
				"title": "CloudWatch Logs",
				"description": "",
				"_type": "tile/data-stream",
				"dataStream": {
					"id": "{{cloud_watch_logs_id}}",
					"pluginConfigId": "{{sample_data_source_id}}"
				},
				"scope": {
					"query": "g.V().order().by('__name').hasNot('__canonicalType').has(\"__configId\", \"{{sample_data_source_id}}\").or(__.has(\"sourceType\", within(\"sample-function\",\"sample-server\",\"sample-database\"))).limit(500)",
					"bindings": {},
					"queryDetail": {}
				}
			}
		},
		{
			"w": 2,
			"h": 3,
			"x": 2,
			"y": 0,
			"i": "a8255dce-5f74-4ff5-b3d3-138f6a0ff130",
			"moved": false,
			"static": false,
			"config": {
				"title": "Lambda Errors",
				"description": "",
				"_type": "tile/data-stream",
				"dataStream": {
					"id": "{{perf_lambda_errors_id}}",
					"pluginConfigId": "{{sample_data_source_id}}",
					"group": {
						"by": [
							"data.lambdaErrors.label",
							"uniqueValues"
						],
						"aggregate": [
							{
								"names": [
									"data.lambdaErrors.value"
								],
								"type": "sum"
							}
						]
					},
					"filter": {
						"filters": [],
						"multiOperation": "and"
					}
				},
				"visualisation": {
					"type": "data-stream-donut-chart"
				},
				"scope": {
					"query": "g.V().order().by('__name').hasNot('__canonicalType').has(\"__configId\", \"{{sample_data_source_id}}\").or(__.has(\"sourceType\", \"sample-function\")).limit(500)",
					"bindings": {},
					"queryDetail": {}
				},
				"monitor": {
					"_type": "simple",
					"tileRollsUp": true,
					"monitorType": "threshold",
					"frequency": 15,
					"aggregation": "top",
					"column": "data.lambdaErrors.value_sum",
					"condition": {
						"columns": [
							"data.lambdaErrors.value_sum"
						],
						"logic": {
							"if": [
								{
									">": [
										{
											"var": "top"
										},
										0
									]
								},
								"error"
							]
						}
					}
				}
			}
		}
	],
	"columns": 4,
	"version": 1
}
EOT
	template_bindings = jsonencode({
	sample_data_source_id = squaredup_datasource.sample_data_source.id
	cloud_watch_logs_id   = local.logs_data_stream.id
	perf_lambda_errors_id = local.perf_lambda_errors_data_stream.id
	})
	workspace_id = squaredup_workspace.application_workspace.id
	display_name = "Sample Dashboard for Workspace Alert Test"
	timeframe    = "last12hours"
}

# Extract ids of tiles
locals {
	dashboard_content = jsondecode(squaredup_dashboard.sample_dashboard.dashboard_content)

	lambda_errors_tile    = [for content in local.dashboard_content.contents : content.i if content.config.title == "Lambda Errors"]
	lambda_errors_tile_id = length(local.lambda_errors_tile) > 0 ? local.lambda_errors_tile[0] : null
}

resource "squaredup_alerting_channel" "slack_api_alert" {
	display_name    = "Slack Alert - Team DevOps for Workspace Alert Test"
	channel_type_id = "channeltype-00000000000000000001"
	config = jsonencode({
	channel = "devops"
	token   = "some-token"
	})
	enabled = true
}

resource "squaredup_workspace_alert" "example" {
	workspace_id = squaredup_workspace.application_workspace.id
	alerting_rules = [
	{
		channel       = squaredup_alerting_channel.slack_api_alert.id
		preview_image = true
		notify_on     = "all_monitors"
	},
	{
		channel       = squaredup_alerting_channel.slack_api_alert.id
		preview_image = true
		notify_on     = "selected_monitors"
		selected_monitors = [
		{
			dashboard_id = squaredup_dashboard.sample_dashboard.id
			tiles_id     = [local.lambda_errors_tile_id]
		}
		]
	}
	]
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_workspace_alert.example", "alerting_rules.#", "2"),
					resource.TestCheckResourceAttr("squaredup_workspace_alert.example", "alerting_rules.1.preview_image", "true"),
				),
			},
		},
	})
}
