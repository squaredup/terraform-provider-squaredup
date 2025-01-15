package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &dataSourceResource{}
	_ resource.ResourceWithConfigure   = &dataSourceResource{}
	_ resource.ResourceWithImportState = &dataSourceResource{}
)

func SquaredupDataSourceResource() resource.Resource {
	return &dataSourceResource{}
}

type dataSourceResource struct {
	client *SquaredUpClient
}

func (r *dataSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasource"
}

type dataSource struct {
	DisplayName  types.String `tfsdk:"display_name"`
	OnPrem       types.Bool   `tfsdk:"on_prem"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"data_source_name"`
	Config       types.String `tfsdk:"config"`
	AgentGroupID types.String `tfsdk:"agent_group_id"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

func (r *dataSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data Sources are used to query third party APIs and SquaredUp visualizes the results",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the data source (Displayed in SquaredUp)",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data source",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_source_name": schema.StringAttribute{
				MarkdownDescription: "Display name of the data source",
				Required:            true,
			},
			"on_prem": schema.BoolAttribute{
				MarkdownDescription: "Whether the data source is an on-prem data source",
				Optional:            true,
				Computed:            true,
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "Sensitive configuration for the data source. Needs to be a valid JSON",
				Optional:            true,
				CustomType:          basetypes.StringType{},
				Sensitive:           true,
			},
			"agent_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the agent group to which the data source should connect to (on-prem data sources only)",
				Optional:            true,
				Computed:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The last time the data source was updated",
				Computed:            true,
			},
		},
	}
}

func (r *dataSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *dataSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataSource
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	var plugin_config map[string]interface{}
	if plan.Config.ValueString() != "" {
		if err := json.Unmarshal([]byte(plan.Config.ValueString()), &plugin_config); err != nil {
			resp.Diagnostics.AddError(
				"Error unmarshalling config",
				fmt.Sprintf("Error unmarshalling config: %v", err),
			)
			return
		}
	}

	newDataSource, err := r.client.AddDataSource(
		plan.DisplayName.ValueString(),
		plan.Name.ValueString(),
		plan.OnPrem.ValueBoolPointer(),
		plugin_config,
		plan.AgentGroupID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating data source",
			fmt.Sprintf("Error creating data source: %v", err),
		)
		return
	}

	state := dataSource{
		DisplayName:  types.StringValue(newDataSource.DisplayName),
		OnPrem:       types.BoolPointerValue(&newDataSource.Plugin.OnPrem),
		Name:         types.StringValue(newDataSource.Plugin.Name),
		AgentGroupID: types.StringValue(newDataSource.AgentGroupID),
		ID:           types.StringValue(newDataSource.ID),
		LastUpdated:  types.StringValue(time.Now().Format(time.RFC850)),
	}

	if plan.Config.ValueString() != "" {
		state.Config = types.StringValue(plan.Config.ValueString())
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *dataSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataSource
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	readDataSource, err := r.client.GetDataSource(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting data source",
			fmt.Sprintf("Error getting data source: %v", err),
		)
		return
	}

	state.DisplayName = types.StringValue(readDataSource.DisplayName)
	state.OnPrem = types.BoolValue(readDataSource.Plugin.OnPrem)
	state.Name = types.StringValue(readDataSource.Plugin.Name)
	state.AgentGroupID = types.StringValue(readDataSource.AgentGroupID)
	state.ID = types.StringValue(readDataSource.ID)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	if state.Config.ValueString() != "" {
		state.Config = types.StringValue(state.Config.ValueString())
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *dataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dataSource
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	var state dataSource
	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	var plugin_config map[string]interface{}
	if plan.Config.ValueString() != "" {
		if err := json.Unmarshal([]byte(plan.Config.ValueString()), &plugin_config); err != nil {
			resp.Diagnostics.AddError(
				"Error unmarshalling config",
				fmt.Sprintf("Error unmarshalling config: %v", err),
			)
			return
		}
	}

	err := r.client.UpdateDataSource(
		state.ID.ValueString(),
		plan.DisplayName.ValueString(),
		plan.Name.ValueString(),
		plan.OnPrem.ValueBoolPointer(),
		plugin_config,
		plan.AgentGroupID.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating data source",
			fmt.Sprintf("Error updating data source: %v", err),
		)
		return
	}

	getDataSource, err := r.client.GetDataSource(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting data source",
			fmt.Sprintf("Error getting data source: %v", err),
		)
		return
	}

	state = dataSource{
		DisplayName:  types.StringValue(getDataSource.DisplayName),
		OnPrem:       types.BoolPointerValue(&getDataSource.Plugin.OnPrem),
		Name:         types.StringValue(getDataSource.Plugin.Name),
		AgentGroupID: types.StringValue(getDataSource.AgentGroupID),
		ID:           types.StringValue(getDataSource.ID),
		LastUpdated:  types.StringValue(time.Now().Format(time.RFC850)),
	}

	if plan.Config.ValueString() != "" {
		state.Config = types.StringValue(plan.Config.ValueString())
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *dataSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dataSource
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteDataSource(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete data source",
			fmt.Sprintf("Error: %v", err),
		)
		return
	}
}

func (r *dataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
