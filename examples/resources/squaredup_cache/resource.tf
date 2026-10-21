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

# Syncs every hour for two objects. Full sync replaces the cached rows on
# every sync.
resource "squaredup_cache" "database_free_space" {
  display_name    = "Database Free Space"
  data_source_id  = squaredup_datasource.sample_data_source.id
  data_stream_id  = data.squaredup_data_streams.database_free_space.data_streams[0].id
  supported_scope = "list"
  object_ids = [
    data.squaredup_nodes.database_1.node_properties[0].id,
    data.squaredup_nodes.database_2.node_properties[0].id,
  ]
  timeframe         = "last24hours"
  frequency_minutes = 60
  sync_mode         = "full"
  retain_mirrors    = 3

  # Destroying a cache permanently deletes its table and data, and tiles
  # using the cache stop working immediately
  lifecycle {
    prevent_destroy = true
  }
}

# Syncs every 15 minutes for every database, selected with a Gremlin query.
# Incremental sync merges new and changed rows into the cached rows. The
# column names are the data stream's column names.
resource "squaredup_cache" "database_free_space_history" {
  display_name    = "Free Space History"
  data_source_id  = squaredup_datasource.sample_data_source.id
  data_stream_id  = data.squaredup_data_streams.database_free_space.data_streams[0].id
  supported_scope = "list"
  scope = jsonencode({
    query       = "g.V().has('sourceType', 'sample-database')"
    bindings    = {}
    queryDetail = {}
  })
  frequency_minutes         = 15
  sync_mode                 = "incremental"
  unique_identifier_columns = ["data.database", "data.timestamp"]
  last_modified_column      = "data.timestamp"
  retention_period_days     = 7
}

# Reference the table name in SQL, for example in the Advanced Query data
# stream on the Cache data source
output "database_free_space_table" {
  value = squaredup_cache.database_free_space.table_name
}
