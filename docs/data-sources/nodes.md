---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "squaredup_nodes Data Source - terraform-provider-squaredup"
subcategory: ""
description: |-
  
---

# squaredup_nodes (Data Source)



## Example Usage

```terraform
data "squaredup_datasources" "sample_data" {
  data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
  display_name     = "Sample Data"
  data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

data "squaredup_nodes" "acommon_node" {
  depends_on     = [squaredup_datasource.sample_data_source]
  data_source_id = squaredup_datasource.sample_data_source.id
  node_name      = "account-common-lambda"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data_source_id` (String) Data Source ID

### Optional

- `node_name` (String) Node Name

### Read-Only

- `node_properties` (Attributes List) Node Properties (see [below for nested schema](#nestedatt--node_properties))

<a id="nestedatt--node_properties"></a>
### Nested Schema for `node_properties`

Read-Only:

- `display_name` (String)
- `id` (String)
- `source_name` (String)
