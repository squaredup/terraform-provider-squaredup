package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourcesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig +
					`
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.squaredup_datasources.sample_data", "plugins.#", "1"),
					resource.TestCheckResourceAttr("data.squaredup_datasources.sample_data", "plugins.0.display_name", "Sample Data"),
				),
			},
		},
	})
}
