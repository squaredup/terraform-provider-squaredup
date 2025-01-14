package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &AlertingChannelResource{}
	_ resource.ResourceWithConfigure   = &AlertingChannelResource{}
	_ resource.ResourceWithImportState = &AlertingChannelResource{}
)

func SquaredUpAlertingChannelResource() resource.Resource {
	return &AlertingChannelResource{}
}

type AlertingChannelResource struct {
	client *SquaredUpClient
}

type squaredupAlertingChannel struct {
	ChannelID     types.String         `tfsdk:"id"`
	DisplayName   types.String         `tfsdk:"display_name"`
	Description   types.String         `tfsdk:"description"`
	ChannelTypeId types.String         `tfsdk:"channel_type_id"`
	Config        jsontypes.Normalized `tfsdk:"config"`
	Enabled       types.Bool           `tfsdk:"enabled"`
	LastUpdated   types.String         `tfsdk:"last_updated"`
}

func (r *AlertingChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerting_channel"
}

func (r *AlertingChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SquaredUp Alerting Channel",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the alerting channel",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the alerting channel",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the alerting channel",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"channel_type_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the alerting channel type",
				Required:            true,
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "The configuration of the alerting channel",
				Required:            true,
				CustomType:          jsontypes.NormalizedType{},
				Sensitive:           true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the alerting channel is enabled",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The last updated time of the alerting channel",
				Computed:            true,
			},
		},
	}
}

func (r *AlertingChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected type for provider data",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *AlertingChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan squaredupAlertingChannel
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	var config map[string]interface{}
	diags = plan.Config.Unmarshal(&config)
	if diags.HasError() {
		return
	}

	// Generate API request body from plan
	alertChannel := AlertingChannel{
		DisplayName:   plan.DisplayName.ValueString(),
		Description:   plan.Description.ValueString(),
		ChannelTypeID: plan.ChannelTypeId.ValueString(),
		Config:        config,
		Enabled:       plan.Enabled.ValueBool(),
	}

	alertingChannel, err := r.client.CreateAlertingChannel(alertChannel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create alerting channel",
			err.Error(),
		)
		return
	}

	state := squaredupAlertingChannel{
		ChannelID:     types.StringValue(alertingChannel.ID),
		DisplayName:   types.StringValue(alertingChannel.DisplayName),
		Description:   types.StringValue(alertingChannel.Description),
		ChannelTypeId: types.StringValue(alertingChannel.ChannelTypeID),
		Config:        plan.Config,
		Enabled:       types.BoolValue(alertingChannel.Enabled),
		LastUpdated:   types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *AlertingChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state squaredupAlertingChannel
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	alertingChannel, err := r.client.GetAlertingChannel(state.ChannelID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get alerting channel",
			err.Error(),
		)
		return
	}

	state = squaredupAlertingChannel{
		ChannelID:     types.StringValue(alertingChannel.ID),
		DisplayName:   types.StringValue(alertingChannel.DisplayName),
		Description:   types.StringValue(alertingChannel.Description),
		ChannelTypeId: types.StringValue(alertingChannel.ChannelTypeID),
		Config:        state.Config,
		Enabled:       types.BoolValue(alertingChannel.Enabled),
		LastUpdated:   types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *AlertingChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan squaredupAlertingChannel
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	var state squaredupAlertingChannel
	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	var config map[string]interface{}
	diags = plan.Config.Unmarshal(&config)
	if diags.HasError() {
		return
	}

	// Generate API request body from plan
	alertChannel := AlertingChannel{
		DisplayName:   plan.DisplayName.ValueString(),
		Description:   plan.Description.ValueString(),
		ChannelTypeID: plan.ChannelTypeId.ValueString(),
		Config:        config,
		Enabled:       plan.Enabled.ValueBool(),
	}

	err := r.client.UpdateAlertingChannel(state.ChannelID.ValueString(), alertChannel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update alerting channel",
			err.Error(),
		)
		return
	}

	readAlertChannel, err := r.client.GetAlertingChannel(state.ChannelID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get alerting channel",
			err.Error(),
		)
		return
	}

	alertingChannel := squaredupAlertingChannel{
		ChannelID:     types.StringValue(readAlertChannel.ID),
		DisplayName:   types.StringValue(readAlertChannel.DisplayName),
		Description:   types.StringValue(readAlertChannel.Description),
		ChannelTypeId: types.StringValue(readAlertChannel.ChannelTypeID),
		Config:        plan.Config,
		Enabled:       types.BoolValue(readAlertChannel.Enabled),
		LastUpdated:   types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &alertingChannel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *AlertingChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state squaredupAlertingChannel
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteAlertingChannel(state.ChannelID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete alerting channel",
			err.Error(),
		)
		return
	}
}

func (r *AlertingChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
