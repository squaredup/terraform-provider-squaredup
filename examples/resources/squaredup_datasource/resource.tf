data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_datasource" "ado_datasource" {
  display_name     = "Azure DevOps"
  data_source_name = "Azure DevOps"
  config = jsonencode({
    org         = "org-name"
    accessToken = "access-token"
  })
}
