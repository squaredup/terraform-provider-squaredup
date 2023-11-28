package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			//Create DataSource Test
			{
				Config: providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - DataSource Test"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_datasource.sample_data_source", "display_name", "Sample Data - DataSource Test"),
					resource.TestCheckResourceAttrSet("squaredup_datasource.sample_data_source", "id"),
					resource.TestCheckResourceAttrSet("squaredup_datasource.sample_data_source", "last_updated"),
				),
			},
			//Update DataSource Test
			{
				Config: providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - DataSource Test Updated"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_datasource.sample_data_source", "display_name", "Sample Data - DataSource Test Updated"),
				),
			},
			// Import DataSource Test
			{
				ResourceName:            "squaredup_datasource.sample_data_source",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
		},
	})
}
