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
