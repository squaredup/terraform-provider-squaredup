package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestNodesDataSource(t *testing.T) {
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
	display_name     = "Sample Data - Nodes Test"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

data "squaredup_nodes" "acommon_node" {
	depends_on = [ squaredup_datasource.sample_data_source ]
	data_source_id = squaredup_datasource.sample_data_source.id
	node_name      = "account-common-lambda"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.squaredup_nodes.sample_data", "node_properties.0.id"),
					resource.TestCheckResourceAttr("data.squaredup_nodes.sample_data", "node_properties.0.display_name", "account-common-lambda"),
				),
			},
		},
	})
}
