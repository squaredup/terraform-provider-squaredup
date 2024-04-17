data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

data "squaredup_nodes" "acommon_node_by_name" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_name      = "account-common-lambda"
}

data "squaredup_nodes" "acommon_node_by_id" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_source_id = "sample-server-2"
}
