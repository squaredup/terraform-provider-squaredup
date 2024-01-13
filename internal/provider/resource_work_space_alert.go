package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &workspaceAlertResource{}
	_ resource.ResourceWithConfigure   = &workspaceAlertResource{}
	_ resource.ResourceWithImportState = &workspaceAlertResource{}
)

func SquaredupWorkspaceAlertResource() resource.Resource {
	return &workspaceAlertResource{}
}

type workspaceAlertResource struct {
	client *SquaredUpClient
}

type workspaceAlerts struct {
	WorkspaceID   types.String     `tfsdk:"workspace_id"`
	AlertingRules []workspaceAlert `tfsdk:"alerting_rules"`
}

type workspaceAlert struct {
	Channel      types.String   `tfsdk:"channel"`
	PreviewImage types.Bool     `tfsdk:"preview_image"`
	Condition    alertCondition `tfsdk:"condition"`
}

type alertCondition struct {
	WorkspaceState  types.Bool     `tfsdk:"workspace_state"`
	IncludeAllTiles types.Bool     `tfsdk:"include_all_tiles"`
	TilesID         []types.String `tfsdk:"tiles_id"`
}

func (r *workspaceAlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace_alert"
}

func (r *workspaceAlertResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SquaredUp Workspace Alert",
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Description: "The ID of the workspace to create the alert in",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alerting_rules": schema.ListNestedAttribute{
				Description: "The alerting rules to create",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"channel": schema.StringAttribute{
							Description: "The ID of the channel to send the alert to",
							Required:    true,
						},
						"preview_image": schema.BoolAttribute{
							Description: "Whether to include a preview image in the alert",
							Default:     booldefault.StaticBool(false),
							Optional:    true,
							Computed:    true,
						},
						"condition": schema.SingleNestedAttribute{
							Description: "The condition for the alert",
							Required:    true,
							Attributes: map[string]schema.Attribute{
								"workspace_state": schema.BoolAttribute{
									Description: "Whether to include the workspace state in the alert",
									Default:     booldefault.StaticBool(false),
									Optional:    true,
									Computed:    true,
								},
								"include_all_tiles": schema.BoolAttribute{
									Description: "Whether to include all tiles in the alert",
									Default:     booldefault.StaticBool(false),
									Optional:    true,
									Computed:    true,
								},
								"tiles_id": schema.ListAttribute{
									Description: "The IDs of the tiles to include in the alert",
									Optional:    true,
									Computed:    true,
									Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *workspaceAlertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *workspaceAlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workspaceAlerts
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

}

func (r *workspaceAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	return
}

func (r *workspaceAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	return
}

func (r *workspaceAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	return
}

func (r *workspaceAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), req, resp)
}
