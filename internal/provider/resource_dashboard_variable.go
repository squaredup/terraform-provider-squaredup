package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DashboardVariableResource{}
	_ resource.ResourceWithConfigure   = &DashboardVariableResource{}
	_ resource.ResourceWithImportState = &DashboardVariableResource{}
)

func SquaredUpDashboardVariableResource() resource.Resource {
	return &DashboardVariableResource{}
}

type DashboardVariableResource struct {
	client *SquaredUpClient
}

type squaredupDashboardVariable struct {
	ID                           types.String `tfsdk:"id"`
	WorkspaceID                  types.String `tfsdk:"workspace_id"`
	CollectionID                 types.String `tfsdk:"collection_id"`
	DefaultObjectSelection       types.String `tfsdk:"default_object_selection"`
	AllowMultipleObjectSelection types.Bool   `tfsdk:"allow_multiple_object_selection"`
	LastUpdated                  types.String `tfsdk:"last_updated"`
}

func (r *DashboardVariableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_variable"
}

func (r *DashboardVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Dashboard variables enable flexible and reusable dashboards by allowing viewers to switch between objects dynamically.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the dashboard variable.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the workspace.",
				Required:            true,
			},
			"collection_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the collection (scope) for the dashboard variable.",
				Required:            true,
			},
			"default_object_selection": schema.StringAttribute{
				MarkdownDescription: "The default object selection for the dashboard variable. Allowed values: `none`, `all`.",
				Required:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					"none",
					"all",
				)},
			},
			"allow_multiple_object_selection": schema.BoolAttribute{
				MarkdownDescription: "Whether to allow multiple object selection for the dashboard variable.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "Last updated timestamp of the dashboard variable",
				Computed:            true,
			},
		},
	}
}

func (r *DashboardVariableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			"Expected SquaredUpClient, got something else.",
		)
		return
	}

	r.client = client
}

func (r *DashboardVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan squaredupDashboardVariable
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if plan.DefaultObjectSelection.ValueString() == "all" && plan.AllowMultipleObjectSelection.ValueBool() {
		resp.Diagnostics.AddError(
			"Invalid configuration: 'allow_multiple_object_selection' cannot be true when 'default_object_selection' is set to 'all'.",
			"'allow_multiple_object_selection' can only be true when 'default_object_selection' is set to 'none'.",
		)
		return
	}

	variable := DashboardVariable{
		Name:                   "Objects",
		Type:                   "object",
		ScopeID:                plan.CollectionID.ValueString(),
		Default:                plan.DefaultObjectSelection.ValueString(),
		AllowMultipleSelection: plan.AllowMultipleObjectSelection.ValueBool(),
	}

	variableRead, err := r.client.CreateDashboardVariable(variable, plan.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create dashboard variable",
			err.Error(),
		)
		return
	}

	state := squaredupDashboardVariable{
		ID:                           types.StringValue(variableRead.ID),
		WorkspaceID:                  types.StringValue(variableRead.WorkspaceID),
		CollectionID:                 types.StringValue(variableRead.Content.ScopeID),
		DefaultObjectSelection:       types.StringValue(variableRead.Content.Default),
		AllowMultipleObjectSelection: types.BoolValue(variableRead.Content.AllowMultipleSelection),
		LastUpdated:                  types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state squaredupDashboardVariable
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	variable, err := r.client.GetDashboardVariable(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read dashboard variable",
			err.Error(),
		)
		return
	}

	state = squaredupDashboardVariable{
		ID:                           types.StringValue(variable.ID),
		WorkspaceID:                  types.StringValue(variable.WorkspaceID),
		CollectionID:                 types.StringValue(variable.Content.ScopeID),
		DefaultObjectSelection:       types.StringValue(variable.Content.Default),
		AllowMultipleObjectSelection: types.BoolValue(variable.Content.AllowMultipleSelection),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan squaredupDashboardVariable
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if plan.DefaultObjectSelection.ValueString() == "all" && plan.AllowMultipleObjectSelection.ValueBool() {
		resp.Diagnostics.AddError(
			"Invalid configuration: 'allow_multiple_object_selection' cannot be true when 'default_object_selection' is set to 'all'.",
			"'allow_multiple_object_selection' can only be true when 'default_object_selection' is set to 'none'.",
		)
		return
	}

	// get dashboard id by performing a read
	variableRead, err := r.client.GetDashboardVariable(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read dashboard variable",
			err.Error(),
		)
		return
	}

	variable := DashboardVariable{
		Name:                   "Objects",
		Type:                   "object",
		ScopeID:                plan.CollectionID.ValueString(),
		Default:                plan.DefaultObjectSelection.ValueString(),
		AllowMultipleSelection: plan.AllowMultipleObjectSelection.ValueBool(),
	}

	if variableRead.Content.DashboardID != "" {
		variable.DashboardID = variableRead.Content.DashboardID
	}

	variableUpdate, err := r.client.UpdateDashboardVariable(plan.ID.ValueString(), variable)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update dashboard variable",
			err.Error(),
		)
		return
	}

	state := squaredupDashboardVariable{
		ID:                           types.StringValue(variableUpdate.ID),
		WorkspaceID:                  types.StringValue(variableUpdate.WorkspaceID),
		CollectionID:                 types.StringValue(variableUpdate.Content.ScopeID),
		AllowMultipleObjectSelection: types.BoolValue(variableUpdate.Content.AllowMultipleSelection),
		DefaultObjectSelection:       types.StringValue(variableUpdate.Content.Default),
		LastUpdated:                  types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state squaredupDashboardVariable
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	err := r.client.DeleteDashboardVariable(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete dashboard variable",
			err.Error(),
		)
		return
	}
}

func (r *DashboardVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
