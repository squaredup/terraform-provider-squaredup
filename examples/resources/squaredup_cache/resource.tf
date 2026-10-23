data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

data "squaredup_data_streams" "database_free_space" {
  data_source_id              = data.squaredup_datasources.sample_data.plugins[0].id
  data_stream_definition_name = "database-free-space"
}

data "squaredup_nodes" "database_1" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_source_id = "sample-database-1"
}

data "squaredup_nodes" "database_2" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_source_id = "sample-database-2"
}

# Syncs every hour for two objects
resource "squaredup_cache" "database_free_space" {
  display_name   = "Database Free Space"
  data_source_id = squaredup_datasource.sample_data_source.id
  data_stream_id = data.squaredup_data_streams.database_free_space.data_streams[0].id
  object_ids = [
    data.squaredup_nodes.database_1.node_properties[0].id,
    data.squaredup_nodes.database_2.node_properties[0].id,
  ]
  timeframe         = "last24hours"
  frequency_minutes = 60

  # Destroying a cache permanently deletes its table and data, and tiles
  # using the cache stop working immediately
  lifecycle {
    prevent_destroy = true
  }
}

# Syncs every 15 minutes for every database, selected with a Gremlin query
resource "squaredup_cache" "all_databases" {
  display_name   = "All Databases"
  data_source_id = squaredup_datasource.sample_data_source.id
  data_stream_id = data.squaredup_data_streams.database_free_space.data_streams[0].id
  scope = jsonencode({
    query       = "g.V().has('sourceType', 'sample-database')"
    bindings    = {}
    queryDetail = {}
  })
  frequency_minutes = 15
}

# Reference the table name in SQL, for example in the Advanced Query data
# stream on the Cache data source
output "database_free_space_table" {
  value = squaredup_cache.database_free_space.table_name
}
