package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DashboardShareResource{}
	_ resource.ResourceWithConfigure   = &DashboardShareResource{}
	_ resource.ResourceWithImportState = &DashboardShareResource{}
)

func SquaredUpDashboardShareResource() resource.Resource {
	return &DashboardShareResource{}
}

type DashboardShareResource struct {
	client *SquaredUpClient
}

type DashboardSharing struct {
	DashboardShareID      types.String `tfsdk:"id"`
	DashboardID           types.String `tfsdk:"dashboard_id"`
	WorkspaceID           types.String `tfsdk:"workspace_id"`
	RequireAuthentication types.Bool   `tfsdk:"require_authentication"`
	EnableLink            types.Bool   `tfsdk:"enabled"`
	LastUpdated           types.String `tfsdk:"last_updated"`
}

func (r *DashboardShareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_share"
}

func (r *DashboardShareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enable sharing for a dashboard",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the dashboard share",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dashboard_id": schema.StringAttribute{
				Description: "The ID of the dashboard to share",
				Required:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "The ID of the workspace where the dashboard is located",
				Required:    true,
			},
			"require_authentication": schema.BoolAttribute{
				Description: "If false, the dashboard will be accessible to anyone with the link",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "If false, sharing of the dashboard is disabled",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"last_updated": schema.StringAttribute{
				Description: "The last time the Dashboard Share was updated",
				Computed:    true,
			},
		},
	}
}

func (r *DashboardShareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DashboardShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DashboardSharing
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	dashboardSharePayload := DashboardShare{
		TargetID:    plan.DashboardID.ValueString(),
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Properties: DashboardShareProperties{
			Enabled:               plan.EnableLink.ValueBool(),
			RequireAuthentication: plan.RequireAuthentication.ValueBool(),
		},
	}

	sharedDashboard, err := r.client.CreateSharedDashboard(dashboardSharePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to share dashboard",
			fmt.Sprintf("Unable to share dashboard: %s", err.Error()),
		)

		return
	}

	state := DashboardSharing{
		DashboardShareID:      types.StringValue(sharedDashboard.ID),
		DashboardID:           types.StringValue(sharedDashboard.TargetID),
		WorkspaceID:           types.StringValue(sharedDashboard.WorkspaceID),
		RequireAuthentication: types.BoolValue(sharedDashboard.Properties.RequireAuthentication),
		EnableLink:            types.BoolValue(sharedDashboard.Properties.Enabled),
		LastUpdated:           types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DashboardSharing
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		return
	}

	sharedDashboard, err := r.client.GetSharedDashboard(state.DashboardShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read shared dashboard",
			fmt.Sprintf("Unable to read shared dashboard: %s", err.Error()),
		)

		return
	}

	state = DashboardSharing{
		DashboardShareID:      types.StringValue(sharedDashboard.ID),
		DashboardID:           types.StringValue(sharedDashboard.TargetID),
		WorkspaceID:           types.StringValue(sharedDashboard.WorkspaceID),
		RequireAuthentication: types.BoolValue(sharedDashboard.Properties.RequireAuthentication),
		EnableLink:            types.BoolValue(sharedDashboard.Properties.Enabled),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DashboardSharing
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	dashboardSharePayload := DashboardShare{
		TargetID:    plan.DashboardID.ValueString(),
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Properties: DashboardShareProperties{
			Enabled:               plan.EnableLink.ValueBool(),
			RequireAuthentication: plan.RequireAuthentication.ValueBool(),
		},
	}

	err := r.client.UpdateSharedDashboard(plan.DashboardShareID.ValueString(), dashboardSharePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update shared dashboard",
			fmt.Sprintf("Unable to update shared dashboard: %s", err.Error()),
		)

		return
	}

	sharedDashboard, err := r.client.GetSharedDashboard(plan.DashboardShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read shared dashboard",
			fmt.Sprintf("Unable to read shared dashboard: %s", err.Error()),
		)

		return
	}

	state := DashboardSharing{
		DashboardShareID:      types.StringValue(sharedDashboard.ID),
		DashboardID:           types.StringValue(sharedDashboard.TargetID),
		WorkspaceID:           types.StringValue(sharedDashboard.WorkspaceID),
		RequireAuthentication: types.BoolValue(sharedDashboard.Properties.RequireAuthentication),
		EnableLink:            types.BoolValue(sharedDashboard.Properties.Enabled),
		LastUpdated:           types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DashboardSharing
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteSharedDashboard(state.DashboardShareID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete shared dashboard",
			fmt.Sprintf("Unable to delete shared dashboard: %s", err.Error()),
		)

		return
	}
}

func (r *DashboardShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
