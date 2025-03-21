package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/pborman/uuid"
)

func TestDataSourceResource(t *testing.T) {
	uuid := uuid.NewRandom().String()
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
	display_name     = "Sample Data - DataSource Test - ` + uuid + `"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_datasource.sample_data_source", "display_name", "Sample Data - DataSource Test - "+uuid),
					resource.TestCheckResourceAttrSet("squaredup_datasource.sample_data_source", "id"),
					resource.TestCheckResourceAttrSet("squaredup_datasource.sample_data_source", "last_updated"),
					resource.TestCheckResourceAttr("squaredup_datasource.sample_data_source", "on_prem", "false"),
				),
			},
			//Update DataSource Test
			{
				Config: providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - DataSource Test Updated - ` + uuid + `"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_datasource.sample_data_source", "display_name", "Sample Data - DataSource Test Updated - "+uuid),
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
