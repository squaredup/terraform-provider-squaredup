package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDashboardResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			//Create Dashboard Test
			{
				Config: providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}
	
resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - Dashboard Test"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}
	
resource "squaredup_workspace" "application_workspace" {
	display_name      = "Application Team - Dashboard Test"
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
			"by": ["data.lambdaErrors.label", "uniqueValues"],
			"aggregate": [
			{
				"names": ["data.lambdaErrors.value"],
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
	display_name = "Sample Dashboard"
	}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard.sample_dashboard", "display_name", "Sample Dashboard"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard.sample_dashboard", "dashboard_content"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard.sample_dashboard", "last_updated"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard.sample_dashboard", "schema_version"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard.sample_dashboard", "id"),
				),
			},
			//Import Dashboard Test
			{
				ResourceName:            "squaredup_dashboard.sample_dashboard",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated", "dashboard_content", "dashboard_template", "template_bindings"},
			},
			//Update Dashboard Test
			{
				Config: providerConfig + `
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
			"by": ["data.lambdaErrors.label", "uniqueValues"],
			"aggregate": [
			{
				"names": ["data.lambdaErrors.value"],
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
	display_name = "Sample Dashboard Updated"
	}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard.sample_dashboard", "display_name", "Sample Dashboard Updated"),
				),
			},
		},
	})
}
