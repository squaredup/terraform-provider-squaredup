---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "squaredup_alerting_channel_types Data Source - squaredup"
subcategory: ""
description: |-
  
---

# squaredup_alerting_channel_types (Data Source)



## Example Usage

```terraform
data "squaredup_alerting_channel_types" "example" {
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `display_name` (String) Filter Alerting Channel Types by Display Name

### Read-Only

- `alerting_channel_types` (Attributes List) Alerting Channel Types are used to configure alert notifications (see [below for nested schema](#nestedatt--alerting_channel_types))

<a id="nestedatt--alerting_channel_types"></a>
### Nested Schema for `alerting_channel_types`

Read-Only:

- `channel_id` (String)
- `description` (String)
- `display_name` (String)
- `image_preview_supported` (Boolean)
- `protocol` (String)
