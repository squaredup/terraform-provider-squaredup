data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "application_workspace" {
  display_name      = "Application Team"
  description       = "Workspace with Dashboards for Application Team"
  datasources_links = [squaredup_datasource.sample_data_source.id]
}

data "squaredup_data_streams" "database_free_space" {
  data_source_id              = data.squaredup_datasources.sample_data.plugins[0].id
  data_stream_definition_name = "database-free-space"
}

# Syncs every hour. Full sync replaces the cached rows on every sync.
resource "squaredup_cache" "database_free_space" {
  display_name    = "Database Free Space"
  data_source_id  = squaredup_datasource.sample_data_source.id
  data_stream_id  = data.squaredup_data_streams.database_free_space.data_streams[0].id
  supported_scope = "list"
  scope = jsonencode({
    workspaceId = squaredup_workspace.application_workspace.id
    query       = "g.V().has('type', 'sample-database')"
  })
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

data "squaredup_data_streams" "logs" {
  data_source_id              = data.squaredup_datasources.sample_data.plugins[0].id
  data_stream_definition_name = "logs"
}

# Syncs every 15 minutes. Incremental sync merges new and changed rows into
# the cached rows.
# Use column names returned by your data stream.
resource "squaredup_cache" "logs" {
  display_name              = "Application Logs"
  data_source_id            = squaredup_datasource.sample_data_source.id
  data_stream_id            = data.squaredup_data_streams.logs.data_streams[0].id
  supported_scope           = "none"
  frequency_minutes         = 15
  sync_mode                 = "incremental"
  unique_identifier_columns = ["id"]
  last_modified_column      = "timestamp"
  retention_period_days     = 7
}

# Reference the table name in SQL, for example in the Advanced Query data
# stream on the Cache data source
output "database_free_space_table" {
  value = squaredup_cache.database_free_space.table_name
}
