package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceScripts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig +
					`
resource "squaredup_script" "example" {
	display_name = "Example Script"
	script_type  = "tileDataJS" #tileDataJS or monitorConditionJS
	script       = <<EOT
async function getData(params, api) {
	// Example: get Star Wars vehicles by making a web request using the '/web/request' api endpoint
	const requestConfig = { method: 'get', url: 'https://swapi.dev/api/vehicles' };
	const vehicles = await api.post('/web/request', requestConfig);

	// Set column metadata
	const columns = ['name', 'model', 'url', 'crew', 'length', 'vehicle_class', 'max_atmosphering_speed', 'passengers', 'films'];
	const metadata = columns.map(c => {
		const column = { name: 'results.' + c };
		if (c === 'url' || c === 'films') {
			column.role = 'link';
			column.shape = 'url';
			column.displayName = c === 'url' ? 'Vehicle Link' : 'Film Link';
		}
		return column;
	});

	// Note: no need to call api.toStreamData when returning data directly from invoking a data stream
	return api.toStreamData(vehicles, { rowPath: ['results', 'films'], metadata } );
}
EOT
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_script.example", "display_name", "Example Script"),
					resource.TestCheckResourceAttr("squaredup_script.example", "script_type", "tileDataJS"),
					resource.TestCheckResourceAttrSet("squaredup_script.example", "id"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_script.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update Test
			{
				Config: providerConfig +
					`
resource "squaredup_script" "example" {
	display_name = "Example Script Updated"
	script_type  = "tileDataJS" #tileDataJS or monitorConditionJS
	script       = <<EOT
async function getData(params, api) {
	// Example: get Star Wars vehicles by making a web request using the '/web/request' api endpoint
	const requestConfig = { method: 'get', url: 'https://swapi.dev/api/vehicles' };
	const vehicles = await api.post('/web/request', requestConfig);

	// Set column metadata
	const columns = ['name', 'model', 'url', 'crew', 'length', 'vehicle_class', 'max_atmosphering_speed', 'passengers', 'films'];
	const metadata = columns.map(c => {
		const column = { name: 'results.' + c };
		if (c === 'url' || c === 'films') {
			column.role = 'link';
			column.shape = 'url';
			column.displayName = c === 'url' ? 'Vehicle Link' : 'Film Link';
		}
		return column;
	});

	// Note: no need to call api.toStreamData when returning data directly from invoking a data stream
	return api.toStreamData(vehicles, { rowPath: ['results', 'films'], metadata } );
}
EOT
}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_script.example", "display_name", "Example Script Updated"),
				),
			},
		},
	})
}
