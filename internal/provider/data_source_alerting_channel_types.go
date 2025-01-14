package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &squaredupAlertingChannelType{}
	_ datasource.DataSourceWithConfigure = &squaredupAlertingChannelType{}
)

func SquaredUpAlertingChannelTypes() datasource.DataSource {
	return &squaredupAlertingChannelType{}
}

type squaredupAlertingChannelType struct {
	client *SquaredUpClient
}

type squaredupAlertingChannelTypes struct {
	AlertingChannelTypes []squaredupAlertingChannelTypeAlertingChannelTypes `tfsdk:"alerting_channel_types"`
	DisplayName          types.String                                       `tfsdk:"display_name"`
}

type squaredupAlertingChannelTypeAlertingChannelTypes struct {
	ChannelID             types.String `tfsdk:"channel_id"`
	DisplayName           types.String `tfsdk:"display_name"`
	Protocol              types.String `tfsdk:"protocol"`
	ImagePreviewSupported types.Bool   `tfsdk:"image_preview_supported"`
	Description           types.String `tfsdk:"description"`
}

func (d *squaredupAlertingChannelType) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerting_channel_types"
}

func (d *squaredupAlertingChannelType) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alerting_channel_types": schema.ListNestedAttribute{
				MarkdownDescription: "Alerting Channel Types are used to configure alert notifications",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"channel_id":              schema.StringAttribute{Computed: true},
						"display_name":            schema.StringAttribute{Computed: true},
						"protocol":                schema.StringAttribute{Computed: true},
						"image_preview_supported": schema.BoolAttribute{Computed: true},
						"description":             schema.StringAttribute{Computed: true},
					},
				},
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter Alerting Channel Types by Display Name",
			},
		},
	}
}

func (d *squaredupAlertingChannelType) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state squaredupAlertingChannelTypes
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertingChannelTypes, err := d.client.GetAlertingChannelTypes(state.DisplayName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get alerting channel types",
			err.Error(),
		)
		return
	}

	for _, alertingChannelType := range alertingChannelTypes {
		alertingChannelTypeState := squaredupAlertingChannelTypeAlertingChannelTypes{
			ChannelID:             types.StringValue(alertingChannelType.ChannelID),
			DisplayName:           types.StringValue(alertingChannelType.DisplayName),
			Protocol:              types.StringValue(alertingChannelType.Protocol),
			ImagePreviewSupported: types.BoolValue(alertingChannelType.ImagePreviewSupported),
			Description:           types.StringValue(alertingChannelType.Description),
		}
		state.AlertingChannelTypes = append(state.AlertingChannelTypes, alertingChannelTypeState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *squaredupAlertingChannelType) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to cast provider data to SquaredUpClient",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}
