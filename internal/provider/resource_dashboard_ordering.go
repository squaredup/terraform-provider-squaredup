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
)

var (
	_ resource.Resource                = &DashboardOrderingResource{}
	_ resource.ResourceWithConfigure   = &DashboardOrderingResource{}
	_ resource.ResourceWithImportState = &DashboardOrderingResource{}
)

func SquaredUpDashboardOrderingResource() resource.Resource {
	return &DashboardOrderingResource{}
}

type DashboardOrderingResource struct {
	client *SquaredUpClient
}

type squaredupDashboardOrdering struct {
	WorkspaceID      types.String `tfsdk:"workspace_id"`
	DashboardIdOrder types.String `tfsdk:"order"`
	ID               types.String `tfsdk:"id"`
	LastUpdated      types.String `tfsdk:"last_updated"`
}

func (r *DashboardOrderingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_ordering"
}

func (r *DashboardOrderingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Specify the order of dashboards and folders on the navigation bar for a given workspace.",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the workspace to manage.",
				Required:            true,
			},
			"order": schema.StringAttribute{
				MarkdownDescription: "The order of the dashboards and folders in the workspace.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the workspace.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The last time the workspace was updated.",
				Computed:            true,
			},
		},
	}
}

func (r *DashboardOrderingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type for provider data",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DashboardOrderingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan squaredupDashboardOrdering
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	workspacePayload, err := BuildWorkspacePayload(plan.DashboardIdOrder.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error building workspace payload",
			fmt.Sprintf("Unable to build workspace payload: %v", err),
		)
		return
	}

	err = r.client.UpdateWorkspace(plan.WorkspaceID.ValueString(), workspacePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workspace order",
			fmt.Sprintf("Unable to create workspace order: %v", err),
		)
		return
	}

	workspace, err := r.client.GetWorkspace(plan.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace order",
			fmt.Sprintf("Unable to read workspace order: %v", err),
		)
		return
	}

	dashboardIdOrderJson, err := json.Marshal(workspace.Data.Properties.DashboardIdOrder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling dashboard ID order",
			fmt.Sprintf("Unable to marshal dashboard ID order: %v", err),
		)
		return
	}

	dashboardIdOrderString := string(dashboardIdOrderJson)

	state := squaredupDashboardOrdering{
		WorkspaceID:      types.StringValue(workspace.ID),
		DashboardIdOrder: types.StringValue(dashboardIdOrderString),
		ID:               types.StringValue(workspace.ID),
		LastUpdated:      types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardOrderingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state squaredupDashboardOrdering
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	workspace, err := r.client.GetWorkspace(state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace order",
			fmt.Sprintf("Unable to read workspace order: %v", err),
		)
		return
	}

	dashboardIdOrderJson, err := json.Marshal(workspace.Data.Properties.DashboardIdOrder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling dashboard ID order",
			fmt.Sprintf("Unable to marshal dashboard ID order: %v", err),
		)
		return
	}

	dashboardIdOrderString := string(dashboardIdOrderJson)

	state = squaredupDashboardOrdering{
		WorkspaceID:      types.StringValue(workspace.ID),
		DashboardIdOrder: types.StringValue(dashboardIdOrderString),
		ID:               types.StringValue(workspace.ID),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardOrderingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan squaredupDashboardOrdering
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	workspacePayload, err := BuildWorkspacePayload(plan.DashboardIdOrder.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error building workspace payload",
			fmt.Sprintf("Unable to build workspace payload: %v", err),
		)
		return
	}

	err = r.client.UpdateWorkspace(plan.WorkspaceID.ValueString(), workspacePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating workspace order",
			fmt.Sprintf("Unable to update workspace order: %v", err),
		)
		return
	}

	readWorkspace, err := r.client.GetWorkspace(plan.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading workspace order",
			fmt.Sprintf("Unable to read workspace order: %v", err),
		)
		return
	}

	dashboardIdOrderJson, err := json.Marshal(readWorkspace.Data.Properties.DashboardIdOrder)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling dashboard ID order",
			fmt.Sprintf("Unable to marshal dashboard ID order: %v", err),
		)
		return
	}

	dashboardIdOrderString := string(dashboardIdOrderJson)

	state := squaredupDashboardOrdering{
		WorkspaceID:      types.StringValue(readWorkspace.ID),
		DashboardIdOrder: types.StringValue(dashboardIdOrderString),
		ID:               types.StringValue(readWorkspace.ID),
		LastUpdated:      types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardOrderingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state squaredupDashboardOrdering
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	dashboardIdOrderPayload := map[string]interface{}{
		"properties": map[string]interface{}{
			"dashboardIdOrder": []interface{}{},
		},
	}

	err := r.client.UpdateWorkspace(state.WorkspaceID.ValueString(), dashboardIdOrderPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting workspace order",
			fmt.Sprintf("Unable to delete workspace order: %v", err),
		)
		return
	}
}

func (r *DashboardOrderingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), req, resp)
}

func BuildWorkspacePayload(dashboardOrderRaw string) (map[string]interface{}, error) {
	var dashboardIdOrder []interface{}
	if err := json.Unmarshal([]byte(dashboardOrderRaw), &dashboardIdOrder); err != nil {
		return nil, fmt.Errorf("unable to parse dashboard ID order: %w", err)
	}

	payload := map[string]interface{}{
		"properties": map[string]interface{}{
			"dashboardIdOrder": dashboardIdOrder,
		},
	}
	return payload, nil
}
