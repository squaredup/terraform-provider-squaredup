resource "squaredup_script" "tileDataJS_example" {
  display_name = "Example Tile Data Script"
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

resource "squaredup_script" "monitorConditionJS_example" {
  display_name = "Example Monitor Condition Script"
  script_type  = "monitorConditionJS" #tileDataJS or monitorConditionJS
  script       = <<EOT
async function getState(params, api) {
    // Get the data rows for the column(s) from which the tile state will be derived.
    // Note: each column in each row has .raw, .value, and .formatted properties.

    const metrics = (await api.getColumnData(params.data, 'metric')).map(row => row.value);

    // A monitor condition script MUST return a state of 'error', 'warning', 'success', or 'unknown',
    // and MAY also return a scalar for the value that caused the state.

    // The following example compares the maximum value (from the 'metric' column) against configured
    // script execution thresholds and sets the state, and scalar that caused the state, accordingly.

    let state = 'unknown';
    const scalar = Math.max(...metrics);

    if (scalar > params.config.errorIfMoreThan) {
        state = 'error';
    } else if (scalar > params.config.warnIfMoreThan) {
        state = 'warning';
    } else if (!Number.isNaN(scalar)) {
        state = 'success';
    }

    return { state, scalar };
}
EOT
}
