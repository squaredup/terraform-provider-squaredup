package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/pborman/uuid"
)

func TestAccSquaredUpScopeResource(t *testing.T) {
	uuid := uuid.NewRandom().String()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - ` + uuid + `"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
	display_name      = "Application Team - ` + uuid + `"
	description       = "Workspace with Dashboards for Application Team"
	datasources_links = [squaredup_datasource.sample_data_source.id]
}

resource "squaredup_scope" "advanced_scope" {
	scope_type     = "advanced"
	display_name   = "Advanced Scope - ` + uuid + `"
	workspace_id   = squaredup_workspace.application_workspace.id
	advanced_query = "g.V().has('__configId', '${squaredup_datasource.sample_data_source.id}').has('sourceId', 'sample-server-2')"
}

data "squaredup_nodes" "acommon_node" {
	depends_on     = [squaredup_datasource.sample_data_source]
	data_source_id = squaredup_datasource.sample_data_source.id
	node_name      = "account-common-lambda"
}

data "squaredup_nodes" "api_node" {
	depends_on     = [squaredup_datasource.sample_data_source]
	data_source_id = squaredup_datasource.sample_data_source.id
	node_name      = "master-api-lambda"
}

resource "squaredup_scope" "fixed_scope" {
	scope_type   = "fixed"
	display_name = "Fixed Scope - ` + uuid + `"
	workspace_id = squaredup_workspace.application_workspace.id
	node_ids     = [data.squaredup_nodes.acommon_node.node_properties[0].id, data.squaredup_nodes.api_node.node_properties[0].id]
}

resource "squaredup_scope" "dynamic_scope" {
	scope_type     = "dynamic"
	display_name   = "Dynamic Scope - ` + uuid + `"
	workspace_id   = squaredup_workspace.application_workspace.id
	data_source_id = [squaredup_datasource.sample_data_source.id]
	types          = ["sample-function"]
	search_query   = "account-common"
}
				  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					//Check Fixed Values
					// Advanced Scope
					resource.TestCheckResourceAttr("squaredup_scope.advanced_scope", "display_name", "Advanced Scope - "+uuid),
					resource.TestCheckResourceAttr("squaredup_scope.advanced_scope", "scope_type", "advanced"),
					// Fixed Scope
					resource.TestCheckResourceAttr("squaredup_scope.fixed_scope", "display_name", "Fixed Scope - "+uuid),
					resource.TestCheckResourceAttr("squaredup_scope.fixed_scope", "scope_type", "fixed"),
					// Dynamic Scope
					resource.TestCheckResourceAttr("squaredup_scope.dynamic_scope", "display_name", "Dynamic Scope - "+uuid),
					resource.TestCheckResourceAttr("squaredup_scope.dynamic_scope", "scope_type", "dynamic"),

					//Check Dynamic Values
					// Advanced Scope
					resource.TestCheckResourceAttrSet("squaredup_scope.advanced_scope", "id"),
					resource.TestCheckResourceAttrSet("squaredup_scope.advanced_scope", "query"),
					// Fixed Scope
					resource.TestCheckResourceAttrSet("squaredup_scope.fixed_scope", "id"),
					resource.TestCheckResourceAttrSet("squaredup_scope.fixed_scope", "node_ids"),
					// Dynamic Scope
					resource.TestCheckResourceAttrSet("squaredup_scope.dynamic_scope", "id"),
					resource.TestCheckResourceAttrSet("squaredup_scope.dynamic_scope", "data_source_id"),
					resource.TestCheckResourceAttrSet("squaredup_scope.dynamic_scope", "types"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_scope.advanced_scope",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			{
				ResourceName:            "squaredup_scope.fixed_scope",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			{
				ResourceName:            "squaredup_scope.dynamic_scope",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update Test
			{
				Config: providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - ` + uuid + `"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
	display_name      = "Application Team - ` + uuid + `"
	description       = "Workspace with Dashboards for Application Team"
	datasources_links = [squaredup_datasource.sample_data_source.id]
}

resource "squaredup_scope" "advanced_scope" {
	scope_type     = "advanced"
	display_name   = "Advanced Scope Updated - ` + uuid + `"
	workspace_id   = squaredup_workspace.application_workspace.id
	advanced_query = "g.V().has('__configId', '${squaredup_datasource.sample_data_source.id}').has('sourceId', 'sample-server-2')"
}

data "squaredup_nodes" "acommon_node" {
	depends_on     = [squaredup_datasource.sample_data_source]
	data_source_id = squaredup_datasource.sample_data_source.id
	node_name      = "account-common-lambda"
}

data "squaredup_nodes" "api_node" {
	depends_on     = [squaredup_datasource.sample_data_source]
	data_source_id = squaredup_datasource.sample_data_source.id
	node_name      = "master-api-lambda"
}

resource "squaredup_scope" "fixed_scope" {
	scope_type   = "fixed"
	display_name = "Fixed Scope Updated - ` + uuid + `"
	workspace_id = squaredup_workspace.application_workspace.id
	node_ids     = [data.squaredup_nodes.acommon_node.node_properties[0].id]
}

resource "squaredup_scope" "dynamic_scope" {
	scope_type     = "dynamic"
	display_name   = "Dynamic Scope Updated - ` + uuid + `"
	workspace_id   = squaredup_workspace.application_workspace.id
	data_source_id = [squaredup_datasource.sample_data_source.id]
	search_query   = "CodePipeline"
}
				  `,
				Check: resource.ComposeAggregateTestCheckFunc(
					//Check Fixed Values
					// Advanced Scope
					resource.TestCheckResourceAttr("squaredup_scope.advanced_scope", "display_name", "Advanced Scope Updated - "+uuid),
					// Fixed Scope
					resource.TestCheckResourceAttr("squaredup_scope.fixed_scope", "display_name", "Fixed Scope Updated - "+uuid),
					resource.TestCheckResourceAttr("squaredup_scope.fixed_scope", "node_ids.#", "1"),
					// Dynamic Scope
					resource.TestCheckResourceAttr("squaredup_scope.dynamic_scope", "display_name", "Dynamic Scope Updated - "+uuid),
					resource.TestCheckResourceAttr("squaredup_scope.dynamic_scope", "search_query", "CodePipeline"),
				),
			},
		},
	})
}
