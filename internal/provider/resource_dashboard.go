package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cbroglie/mustache"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DashboardResource{}
	_ resource.ResourceWithConfigure   = &DashboardResource{}
	_ resource.ResourceWithImportState = &DashboardResource{}
)

func SquaredUpDashboardResource() resource.Resource {
	return &DashboardResource{}
}

type DashboardResource struct {
	client *SquaredUpClient
}

type squaredupDashboard struct {
	DashboardID       types.String         `tfsdk:"id"`
	DisplayName       types.String         `tfsdk:"display_name"`
	WorkspaceID       types.String         `tfsdk:"workspace_id"`
	DashboardTemplate types.String         `tfsdk:"dashboard_template"`
	DashboardVariable types.String         `tfsdk:"dashboard_variable_id"`
	TemplateBindings  jsontypes.Normalized `tfsdk:"template_bindings"`
	DashboardContent  jsontypes.Normalized `tfsdk:"dashboard_content"`
	Timeframe         types.String         `tfsdk:"timeframe"`
	SchemaVersion     types.String         `tfsdk:"schema_version"`
	LastUpdated       types.String         `tfsdk:"last_updated"`
}

func (r *DashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *DashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Dashboard are used to visualize data from Data Sources",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the dashboard",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the dashboard",
				Required:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the workspace where the dashboard is located",
				Required:            true,
			},
			"dashboard_template": schema.StringAttribute{
				MarkdownDescription: "Dashboard template to use for the dashboard",
				Required:            true,
			},
			"dashboard_variable_id": schema.StringAttribute{
				MarkdownDescription: "ID of the dashboard variable to use for this dashboard",
				Optional:            true,
				Computed:            true,
			},
			"template_bindings": schema.StringAttribute{
				MarkdownDescription: "Template Bindings used for replacing mustache template in the dashboard template. Needs to be a JSON encoded string.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"dashboard_content": schema.StringAttribute{
				MarkdownDescription: "The content of the dashboard. This is the rendered dashboard template with the template bindings applied.",
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"timeframe": schema.StringAttribute{
				MarkdownDescription: "The timeframe of the dashboard. It should be one of the following: last1hour, last12hours, last24hours, last7days, last30days, thisMonth, thisQuarter, thisYear, lastMonth, lastQuarter, lastYear",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					"last1hour",
					"last12hours",
					"last24hours",
					"last7days",
					"last30days",
					"thisMonth",
					"thisQuarter",
					"thisYear",
					"lastMonth",
					"lastQuarter",
					"lastYear",
				)},
			},
			"schema_version": schema.StringAttribute{
				MarkdownDescription: "The schema version of the dashboard",
				Optional:            true,
				Computed:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The last updated date of the dashboard",
				Computed:            true,
			},
		},
	}
}

func (r *DashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan squaredupDashboard
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	var templateBindings map[string]interface{}
	var updatedDashboard string

	if plan.TemplateBindings.ValueString() != "" {
		diags = plan.TemplateBindings.Unmarshal(&templateBindings)
		if diags.HasError() {
			return
		}

		updatedTemplate, err := mustache.Render(plan.DashboardTemplate.ValueString(), templateBindings)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to render template",
				err.Error(),
			)
			return
		}
		// Check if the rendered template is valid JSON
		var jsonData interface{}
		err = json.Unmarshal([]byte(updatedTemplate), &jsonData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Rendered template is not a valid JSON. Please check that the JSON is valid after rendering the template with the template bindings.",
				err.Error(),
			)
			return
		}
		updatedDashboard = updatedTemplate
	} else {
		updatedDashboard = plan.DashboardTemplate.ValueString()
		plan.TemplateBindings = jsontypes.NewNormalizedNull()
	}

	dashboard, err := r.client.CreateDashboard(plan.DisplayName.ValueString(), plan.WorkspaceID.ValueString(), plan.Timeframe.ValueString(), updatedDashboard)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create dashboard",
			err.Error(),
		)
		return
	}

	var dashboardVariableID string
	if plan.DashboardVariable.ValueString() != "" {
		dashboardVariableID, err = UpdateDashboardVariable(r, dashboard.ID, plan.DashboardVariable.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update dashboard variable",
				err.Error(),
			)
			return
		}
	}

	state := squaredupDashboard{
		DashboardID:       types.StringValue(dashboard.ID),
		DisplayName:       types.StringValue(dashboard.DisplayName),
		WorkspaceID:       types.StringValue(dashboard.WorkspaceID),
		DashboardTemplate: plan.DashboardTemplate,
		TemplateBindings:  plan.TemplateBindings,
		DashboardContent:  jsontypes.NewNormalizedValue(updatedDashboard),
		Timeframe:         types.StringValue(dashboard.Timeframe),
		SchemaVersion:     types.StringValue(dashboard.SchemaVersion),
		LastUpdated:       types.StringValue(time.Now().Format(time.RFC850)),
	}

	if dashboardVariableID != "" {
		state.DashboardVariable = types.StringValue(dashboardVariableID)
	} else {
		state.DashboardVariable = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state squaredupDashboard
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	dashboard, err := r.client.GetDashboard(state.DashboardID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get dashboard",
			err.Error(),
		)
		return
	}

	state = squaredupDashboard{
		DashboardID:       types.StringValue(dashboard.ID),
		DisplayName:       types.StringValue(dashboard.DisplayName),
		WorkspaceID:       types.StringValue(dashboard.WorkspaceID),
		DashboardTemplate: state.DashboardTemplate,
		TemplateBindings:  state.TemplateBindings,
		DashboardContent:  state.DashboardContent,
		Timeframe:         types.StringValue(dashboard.Timeframe),
		SchemaVersion:     types.StringValue(dashboard.SchemaVersion),
	}

	// Check if the dashboard variable ID is set
	if state.DashboardVariable.ValueString() != "" {
		dashboardVariable, err := r.client.GetDashboardVariable(state.DashboardVariable.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get dashboard variable",
				err.Error(),
			)
			return
		}
		state.DashboardVariable = types.StringValue(dashboardVariable.ID)
	} else {
		state.DashboardVariable = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan squaredupDashboard
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	var templateBindings map[string]interface{}
	var updatedDashboard string

	if plan.TemplateBindings.ValueString() != "" {
		diags = plan.TemplateBindings.Unmarshal(&templateBindings)
		if diags.HasError() {
			return
		}

		updatedTemplate, err := mustache.Render(plan.DashboardTemplate.ValueString(), templateBindings)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to render template",
				err.Error(),
			)
			return
		}
		// Check if the rendered template is valid JSON
		var jsonData interface{}
		err = json.Unmarshal([]byte(updatedTemplate), &jsonData)
		if err != nil {
			resp.Diagnostics.AddError(
				"Rendered template is not a valid JSON. Please check that the JSON is valid after rendering the template with the template bindings.",
				err.Error(),
			)
			return
		}
		updatedDashboard = updatedTemplate
	} else {
		updatedDashboard = plan.DashboardTemplate.ValueString()
		plan.TemplateBindings = jsontypes.NewNormalizedNull()
	}

	dashboard, err := r.client.UpdateDashboard(plan.DashboardID.ValueString(), plan.DisplayName.ValueString(), plan.Timeframe.ValueString(), updatedDashboard)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update dashboard",
			err.Error(),
		)
		return
	}

	var dashboardVariableID string
	if plan.DashboardVariable.ValueString() != "" {
		dashboardVariableID, err = UpdateDashboardVariable(r, dashboard.ID, plan.DashboardVariable.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update dashboard variable",
				err.Error(),
			)
			return
		}
	} else {
		dashboardVariableID = ""
	}

	plan = squaredupDashboard{
		DashboardID:       types.StringValue(dashboard.ID),
		DisplayName:       types.StringValue(dashboard.DisplayName),
		WorkspaceID:       types.StringValue(dashboard.WorkspaceID),
		DashboardTemplate: plan.DashboardTemplate,
		TemplateBindings:  plan.TemplateBindings,
		DashboardContent:  jsontypes.NewNormalizedValue(updatedDashboard),
		Timeframe:         types.StringValue(dashboard.Timeframe),
		SchemaVersion:     types.StringValue(dashboard.SchemaVersion),
		LastUpdated:       types.StringValue(time.Now().Format(time.RFC850)),
	}

	if dashboardVariableID != "" {
		plan.DashboardVariable = types.StringValue(dashboardVariableID)
	} else {
		plan.DashboardVariable = types.StringNull()
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state squaredupDashboard
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteDashboard(state.DashboardID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete dashboard",
			err.Error(),
		)
		return
	}
}

func (r *DashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func UpdateDashboardVariable(squaredupProvider *DashboardResource, dashboardID string, variableId string) (string, error) {
	dashboardVariable, err := squaredupProvider.client.GetDashboardVariable(variableId)
	if err != nil {
		return "", err
	}

	updateRequestBody := DashboardVariable{
		Name:                   dashboardVariable.Name,
		Type:                   dashboardVariable.Content.Type,
		ScopeID:                dashboardVariable.Content.ScopeID,
		Default:                dashboardVariable.Content.Default,
		AllowMultipleSelection: dashboardVariable.Content.AllowMultipleSelection,
		DashboardID:            dashboardID,
	}

	updatedDashboardVariable, err := squaredupProvider.client.UpdateDashboardVariable(dashboardVariable.ID, updateRequestBody)
	if err != nil {
		return "", err
	}

	return updatedDashboardVariable.ID, nil
}
