package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	Channel          types.String       `tfsdk:"channel"`
	PreviewImage     types.Bool         `tfsdk:"preview_image"`
	NotifyOn         types.String       `tfsdk:"notify_on"`
	SelectedMonitors []SelectedMonitors `tfsdk:"selected_monitors"`
}

type SelectedMonitors struct {
	DashboardID types.String   `tfsdk:"dashboard_id"`
	TilesID     []types.String `tfsdk:"tiles_id"`
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
						"notify_on": schema.StringAttribute{
							Description: "Condition to trigger the alert. Must be one of: 'workspace_state', 'all_monitors', or 'selected_monitors'",
							Required:    true,
							Validators: []validator.String{stringvalidator.OneOf(
								"workspace_state",
								"all_monitors",
								"selected_monitors",
							)},
						},
						"selected_monitors": schema.ListNestedAttribute{
							Description: "The monitors to trigger the alert on. Required if notify_on is 'selected_monitors'",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"dashboard_id": schema.StringAttribute{
										Description: "The ID of the dashboard where the monitor is configured",
										Required:    true,
									},
									"tiles_id": schema.ListAttribute{
										Description: "The ID of the tiles to trigger the alert on",
										Required:    true,
										ElementType: types.StringType,
									},
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

	payload, err, warning := constructPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error constructing JSON", err.Error())
		return
	}

	if warning != "" {
		resp.Diagnostics.AddWarning("Unsupported Attribute", warning)
	}

	fmt.Printf("JSON: %s\n", payload)

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

func constructPayload(plan workspaceAlerts) (string, error, string) {
	var result WorkspaceAlertsData
	var warning string

	for _, rule := range plan.AlertingRules {
		var channels []AlertChannel

		channel := AlertChannel{
			ID:                  rule.Channel.ValueString(),
			IncludePreviewImage: rule.PreviewImage.ValueBool(),
		}

		if rule.NotifyOn.ValueString() == "workspace_state" && rule.PreviewImage.ValueBool() == true {
			channel.IncludePreviewImage = false
			warning = "Preview images are not supported when using 'workspace_state' for 'notify_on'. The 'preview_image' attribute will be ignored."
		}

		channels = append(channels, channel)

		var conditions AlertConditions
		conditions.Monitors.IncludeAllTiles = rule.NotifyOn.ValueString() == "all_monitors"
		conditions.Monitors.DashboardRollupHealth = false
		conditions.Monitors.RollupHealth = false

		if rule.NotifyOn.ValueString() == "workspace_state" {
			conditions.Monitors.RollupHealth = true
		}

		if rule.NotifyOn.ValueString() == "selected_monitors" {
			conditions.Monitors.Dashboards = make(map[string]AlertDashboard)
			for _, selectedMonitor := range rule.SelectedMonitors {
				dashboardID := selectedMonitor.DashboardID.ValueString()
				dashboard := AlertDashboard{
					Tiles: make(map[string]AlertTile),
				}

				for _, tileID := range selectedMonitor.TilesID {
					dashboard.Tiles[tileID.ValueString()] = AlertTile{
						Include: true,
					}
				}

				conditions.Monitors.Dashboards[dashboardID] = dashboard
			}
		}

		result.AlertingRules = append(result.AlertingRules, WorkspaceAlertData{
			Channels:   channels,
			Conditions: conditions,
		})
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err, ""
	}

	return string(jsonData), nil, warning
}
