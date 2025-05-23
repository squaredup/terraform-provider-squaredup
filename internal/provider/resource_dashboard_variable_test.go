package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/pborman/uuid"
)

func TestAccSquaredUpDashboardVariableResource(t *testing.T) {
	uniqueId := uuid.NewRandom().String()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + fmt.Sprintf(`
data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data - %s"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
  display_name      = "Application Team - %s"
  description       = "Workspace with Dashboards for Application Team"
  datasources_links = [squaredup_datasource.sample_data_source.id]
}

resource "squaredup_scope" "dynamic_scope" {
  scope_type     = "dynamic"
  display_name   = "Dynamic Scope - %s"
  workspace_id   = squaredup_workspace.application_workspace.id
  data_source_id = [squaredup_datasource.sample_data_source.id]
  types          = ["sample-function"]
  search_query   = "account-common"
}

resource "squaredup_dashboard_variable" "example_all_variable" {
  workspace_id             = squaredup_workspace.application_workspace.id
  collection_id            = squaredup_scope.dynamic_scope.id
  default_object_selection = "all"
}

resource "squaredup_dashboard_variable" "example_none_variable" {
  workspace_id                      = squaredup_workspace.application_workspace.id
  collection_id                     = squaredup_scope.dynamic_scope.id
  default_object_selection          = "none"
  allow_multiple_object_selection   = true
}
`, uniqueId, uniqueId, uniqueId),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check "all" variable
					resource.TestCheckResourceAttr("squaredup_dashboard_variable.example_all_variable", "default_object_selection", "all"),
					resource.TestCheckResourceAttr("squaredup_dashboard_variable.example_all_variable", "allow_multiple_object_selection", "false"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard_variable.example_all_variable", "id"),
					// Check "none" variable
					resource.TestCheckResourceAttr("squaredup_dashboard_variable.example_none_variable", "default_object_selection", "none"),
					resource.TestCheckResourceAttr("squaredup_dashboard_variable.example_none_variable", "allow_multiple_object_selection", "true"),
					resource.TestCheckResourceAttrSet("squaredup_dashboard_variable.example_none_variable", "id"),
				),
			},
			// Import Test for "all" variable
			{
				ResourceName:      "squaredup_dashboard_variable.example_all_variable",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["squaredup_dashboard_variable.example_all_variable"].Primary.ID, nil
				},
			},
			// Import Test for "none" variable
			{
				ResourceName:      "squaredup_dashboard_variable.example_none_variable",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["squaredup_dashboard_variable.example_none_variable"].Primary.ID, nil
				},
			},
			// Update Test
			{
				Config: providerConfig + fmt.Sprintf(`
data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data - %s"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
  display_name      = "Application Team - %s"
  description       = "Workspace with Dashboards for Application Team"
  datasources_links = [squaredup_datasource.sample_data_source.id]
}

resource "squaredup_scope" "dynamic_scope" {
  scope_type     = "dynamic"
  display_name   = "Dynamic Scope - %s"
  workspace_id   = squaredup_workspace.application_workspace.id
  data_source_id = [squaredup_datasource.sample_data_source.id]
  types          = ["sample-function"]
  search_query   = "account-common"
}

resource "squaredup_dashboard_variable" "example_all_variable" {
  workspace_id             = squaredup_workspace.application_workspace.id
  collection_id            = squaredup_scope.dynamic_scope.id
  default_object_selection = "none"
}

resource "squaredup_dashboard_variable" "example_none_variable" {
  workspace_id                      = squaredup_workspace.application_workspace.id
  collection_id                     = squaredup_scope.dynamic_scope.id
  default_object_selection          = "none"
  allow_multiple_object_selection   = false
}
`, uniqueId, uniqueId, uniqueId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_dashboard_variable.example_all_variable", "default_object_selection", "none"),
					resource.TestCheckResourceAttr("squaredup_dashboard_variable.example_none_variable", "allow_multiple_object_selection", "false"),
				),
			},
		},
	})
}
