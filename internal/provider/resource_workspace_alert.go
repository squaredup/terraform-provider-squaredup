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
	ID            types.String     `tfsdk:"id"`
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
			"id": schema.StringAttribute{
				Description: "The ID of the workspace",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

	err = r.client.UpdateWorkspace(plan.WorkspaceID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workspace alerts", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.WorkspaceID.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *workspaceAlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspaceAlerts
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	readWorkspace, err := r.client.GetWorkspace(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading workspace", err.Error())
		return
	}

	alertingRules, err := constructAlertingRules(readWorkspace)
	if err != nil {
		resp.Diagnostics.AddError("Error constructing alerting rules", err.Error())
		return
	}

	updatedState := workspaceAlerts{
		WorkspaceID:   state.ID,
		AlertingRules: alertingRules,
		ID:            state.ID,
	}

	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceAlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	err = r.client.UpdateWorkspace(plan.WorkspaceID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workspace alerts", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.WorkspaceID.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceAlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workspaceAlerts
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	payload := map[string]interface{}{
		"alertingRules": []interface{}{},
	}

	err := r.client.UpdateWorkspace(state.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error with removing workspace alerts", err.Error())
		return
	}
}

func (r *workspaceAlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func constructPayload(plan workspaceAlerts) (map[string]interface{}, error, string) {
	var result WorkspaceAlertsData
	var warning string

	for _, rule := range plan.AlertingRules {
		var channels []AlertChannel

		channel := AlertChannel{
			ID:                  rule.Channel.ValueString(),
			IncludePreviewImage: rule.PreviewImage.ValueBool(),
		}

		if rule.NotifyOn.ValueString() == "workspace_state" && rule.PreviewImage.ValueBool() {
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
		return nil, err, ""
	}

	var jsonDataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonDataMap); err != nil {
		return nil, err, ""
	}

	return jsonDataMap, nil, warning
}

func determineNotifyOn(monitors AlertMonitors) (string, error) {
	if monitors.IncludeAllTiles {
		return "all_monitors", nil
	} else if monitors.RollupHealth {
		return "workspace_state", nil
	} else if len(monitors.Dashboards) >= 0 {
		return "selected_monitors", nil
	}

	err := fmt.Errorf("unable to determine notify_on value")
	return "", err
}

func constructAlertingRules(readWorkspaceData *WorkspaceRead) ([]workspaceAlert, error) {
	var alertingRules []workspaceAlert

	for _, rule := range readWorkspaceData.Data.AlertingRules {
		var selectedMonitors []SelectedMonitors
		for dashID, dashTiles := range rule.Conditions.Monitors.Dashboards {
			var tilesIDs []types.String
			for tileID, tile := range dashTiles.Tiles {
				if tile.Include {
					tilesIDs = append(tilesIDs, types.StringValue(tileID))
				}
			}
			selectedMonitors = append(selectedMonitors, SelectedMonitors{
				DashboardID: types.StringValue(dashID),
				TilesID:     tilesIDs,
			})
		}

		notifyOn, err := determineNotifyOn(rule.Conditions.Monitors)
		if err != nil {
			return nil, err
		}

		alertingRule := workspaceAlert{
			Channel:          types.StringValue(rule.Channels[0].ID),
			PreviewImage:     types.BoolValue(rule.Channels[0].IncludePreviewImage),
			NotifyOn:         types.StringValue(notifyOn),
			SelectedMonitors: selectedMonitors,
		}
		alertingRules = append(alertingRules, alertingRule)
	}

	return alertingRules, nil
}
