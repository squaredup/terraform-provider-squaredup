package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataStreamsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig +
					`
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

data "squaredup_data_streams" "sample_data_logs_dataStreams" {
	data_source_id = data.squaredup_datasources.sample_data.plugins[0].id
	data_stream_definition_name = "logs"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.squaredup_data_streams.sample_data_logs_dataStreams", "data_streams.#", "1"),
					resource.TestCheckResourceAttr("data.squaredup_data_streams.sample_data_logs_dataStreams", "data_streams.0.definition_name", "logs"),
				),
			},
		},
	})
}
