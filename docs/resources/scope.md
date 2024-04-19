---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "squaredup_scope Resource - terraform-provider-squaredup"
subcategory: ""
description: |-
  SquaredUp Scope
---

# squaredup_scope (Resource)

SquaredUp Scope

## Example Usage

```terraform
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

resource "squaredup_scope" "advanced_scope" {
  scope_type     = "advanced"
  display_name   = "Advanced Scope"
  workspace_id   = squaredup_workspace.application_workspace.id
  advanced_query = "g.V().has('__configId', '${squaredup_datasource.sample_data_source.id}').has('sourceId', 'sample-server-2')" //any gremlin query
}

data "squaredup_nodes" "acommon_node" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_name      = "account-common-lambda"
}

data "squaredup_nodes" "api_node" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_name      = "master-api-lambda"
}

resource "squaredup_scope" "fixed_scope" {
  scope_type   = "fixed"
  display_name = "Fixed Scope"
  workspace_id = squaredup_workspace.application_workspace.id
  node_ids     = [data.squaredup_nodes.acommon_node.node_properties[0].id, data.squaredup_nodes.api_node.node_properties[0].id]
}

resource "squaredup_scope" "dynamic_scope" {
  scope_type     = "dynamic"
  display_name   = "Dynamic Scope"
  workspace_id   = squaredup_workspace.application_workspace.id
  data_source_id = [squaredup_datasource.sample_data_source.id] //if no data source is provided, it will search within all
  types          = ["sample-function"]                          //if no type is provided, it will search within all
  search_query   = "account-common"                             //similar to search bar
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `display_name` (String) Display name for the scope
- `scope_type` (String) Type of the scope. Either 'dynamic' or 'fixed'
- `workspace_id` (String) ID of the workspace

### Optional

- `advanced_query` (String) Advanced query (Gremlin)
- `data_source_id` (List of String) IDs of the data sources to filter the scope
- `node_ids` (List of String) IDs of the nodes that scope will contain
- `search_query` (String) Search query
- `types` (List of String) Node types to filter the scope

### Read-Only

- `id` (String) ID of the scope
- `last_updated` (String) Last updated timestamp
- `query` (String) Query for the scope

## Import

Import is supported using the following syntax:

```shell
# Scopes can be imported by specifying workspace id and scope id
terraform import squaredup_scope.example space-123,scope-123
```