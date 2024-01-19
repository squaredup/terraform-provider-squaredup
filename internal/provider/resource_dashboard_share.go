package provider

import (
	"context"
	"fmt"
	"strings"
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
	_ resource.Resource                = &OpenAccessResource{}
	_ resource.ResourceWithConfigure   = &OpenAccessResource{}
	_ resource.ResourceWithImportState = &OpenAccessResource{}
)

func SquaredUpOpenAccessResource() resource.Resource {
	return &OpenAccessResource{}
}

type OpenAccessResource struct {
	client *SquaredUpClient
}

type DashboardSharing struct {
	OpenAccessID          types.String `tfsdk:"id"`
	DashboardID           types.String `tfsdk:"dashboard_id"`
	WorkspaceID           types.String `tfsdk:"workspace_id"`
	RequireAuthentication types.Bool   `tfsdk:"require_authentication"`
	EnableLink            types.Bool   `tfsdk:"enabled"`
	OpenAccessLink        types.String `tfsdk:"dashboard_share_link"`
	LastUpdated           types.String `tfsdk:"last_updated"`
}

func (r *OpenAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_share"
}

func (r *OpenAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Enable Open Access for a dashboard",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Shared dashboard",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dashboard_id": schema.StringAttribute{
				Description: "The ID of the dashboard",
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
			"dashboard_share_link": schema.StringAttribute{
				Description: "The Open Access Link for the dashboard",
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "The last time the Open Access was updated",
				Computed:    true,
			},
		},
	}
}

func (r *OpenAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OpenAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DashboardSharing
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	OpenAccessPayload := OpenAccess{
		TargetID:    plan.DashboardID.ValueString(),
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Properties: OpenAccessProperties{
			Enabled:               plan.EnableLink.ValueBool(),
			RequireAuthentication: plan.RequireAuthentication.ValueBool(),
		},
	}

	openAccess, err := r.client.CreateOpenAccess(OpenAccessPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to share dashboard",
			fmt.Sprintf("Unable to share dashboard: %s", err.Error()),
		)

		return
	}

	state := DashboardSharing{
		OpenAccessID:          types.StringValue(openAccess.ID),
		DashboardID:           types.StringValue(openAccess.TargetID),
		WorkspaceID:           types.StringValue(openAccess.WorkspaceID),
		RequireAuthentication: types.BoolValue(openAccess.Properties.RequireAuthentication),
		EnableLink:            types.BoolValue(openAccess.Properties.Enabled),
		OpenAccessLink:        types.StringValue(generateOpenAccessURL(openAccess.ID)),
		LastUpdated:           types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *OpenAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DashboardSharing
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		return
	}

	openAccess, err := r.client.GetOpenAccess(state.OpenAccessID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read shared dashboard",
			fmt.Sprintf("Unable to read shared dashboard: %s", err.Error()),
		)

		return
	}

	state = DashboardSharing{
		OpenAccessID:          types.StringValue(openAccess.ID),
		DashboardID:           types.StringValue(openAccess.TargetID),
		WorkspaceID:           types.StringValue(openAccess.WorkspaceID),
		RequireAuthentication: types.BoolValue(openAccess.Properties.RequireAuthentication),
		EnableLink:            types.BoolValue(openAccess.Properties.Enabled),
		OpenAccessLink:        types.StringValue(generateOpenAccessURL(openAccess.ID)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *OpenAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DashboardSharing
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	OpenAccessPayload := OpenAccess{
		TargetID:    plan.DashboardID.ValueString(),
		WorkspaceID: plan.WorkspaceID.ValueString(),
		Properties: OpenAccessProperties{
			Enabled:               plan.EnableLink.ValueBool(),
			RequireAuthentication: plan.RequireAuthentication.ValueBool(),
		},
	}

	err := r.client.UpdateOpenAccess(plan.OpenAccessID.ValueString(), OpenAccessPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update shared dashboard",
			fmt.Sprintf("Unable to update shared dashboard: %s", err.Error()),
		)

		return
	}

	readOpenAccess, err := r.client.GetOpenAccess(plan.OpenAccessID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read shared dashboard",
			fmt.Sprintf("Unable to read shared dashboard: %s", err.Error()),
		)

		return
	}

	state := DashboardSharing{
		OpenAccessID:          types.StringValue(readOpenAccess.ID),
		DashboardID:           types.StringValue(readOpenAccess.TargetID),
		WorkspaceID:           types.StringValue(readOpenAccess.WorkspaceID),
		RequireAuthentication: types.BoolValue(readOpenAccess.Properties.RequireAuthentication),
		EnableLink:            types.BoolValue(readOpenAccess.Properties.Enabled),
		OpenAccessLink:        types.StringValue(generateOpenAccessURL(readOpenAccess.ID)),
		LastUpdated:           types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *OpenAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DashboardSharing
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteOpenAccess(state.OpenAccessID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete shared dashboard",
			fmt.Sprintf("Unable to delete shared dashboard: %s", err.Error()),
		)

		return
	}
}

func (r *OpenAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func generateOpenAccessURL(id string) string {
	return "https://app.squaredup.com/openaccess/" + strings.Split(id, "-")[1]
}
