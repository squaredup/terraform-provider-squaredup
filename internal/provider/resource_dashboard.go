package provider

import (
	"context"
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
	DashboardTemplate jsontypes.Normalized `tfsdk:"dashboard_template"`
	TemplateBindings  jsontypes.Normalized `tfsdk:"template_bindings"`
	DashboardContent  jsontypes.Normalized `tfsdk:"dashboard_content"`
	Timeframe         types.String         `tfsdk:"timeframe"`
	Name              types.String         `tfsdk:"name"`
	SchemaVersion     types.String         `tfsdk:"schema_version"`
	LastUpdated       types.String         `tfsdk:"last_updated"`
}

func (r *DashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *DashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Dashboard are used to visualize data from Data Sources",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the dashboard",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the dashboard",
				Required:    true,
			},
			"workspace_id": schema.StringAttribute{
				Description: "The ID of the workspace where the dashboard is located",
				Required:    true,
			},
			"dashboard_template": schema.StringAttribute{
				Description: "Dashboard template to use for the dashboard",
				Required:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"template_bindings": schema.StringAttribute{
				Description: "Template Bindings used for replacing mustache template in the dashboard template. Needs to be a JSON encoded string.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"dashboard_content": schema.StringAttribute{
				Description: "The content of the dashboard. This is the rendered dashboard template with the template bindings applied.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"timeframe": schema.StringAttribute{
				Description: "The timeframe of the dashboard. It should be one of the following: last1hour, last12hours, last24hours, last7days, last30days, thisMonth, thisQuarter, thisYear, lastMonth, lastQuarter, lastYear",
				Optional:    true,
				Computed:    true,
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
			"name": schema.StringAttribute{
				Description: "The name of the dashboard",
				Computed:    true,
			},
			"schema_version": schema.StringAttribute{
				Description: "The schema version of the dashboard",
				Optional:    true,
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "The last updated date of the dashboard",
				Computed:    true,
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

	state := squaredupDashboard{
		DashboardID:       types.StringValue(dashboard.ID),
		DisplayName:       types.StringValue(dashboard.DisplayName),
		WorkspaceID:       types.StringValue(dashboard.WorkspaceID),
		DashboardTemplate: plan.DashboardTemplate,
		TemplateBindings:  plan.TemplateBindings,
		DashboardContent:  jsontypes.NewNormalizedValue(updatedDashboard),
		Timeframe:         types.StringValue(dashboard.Timeframe),
		Name:              types.StringValue(dashboard.Name),
		SchemaVersion:     types.StringValue(dashboard.SchemaVersion),
		LastUpdated:       types.StringValue(time.Now().Format(time.RFC850)),
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
		Name:              types.StringValue(dashboard.Name),
		SchemaVersion:     types.StringValue(dashboard.SchemaVersion),
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

	plan = squaredupDashboard{
		DashboardID:       types.StringValue(dashboard.ID),
		DisplayName:       types.StringValue(dashboard.DisplayName),
		WorkspaceID:       types.StringValue(dashboard.WorkspaceID),
		DashboardTemplate: plan.DashboardTemplate,
		TemplateBindings:  plan.TemplateBindings,
		DashboardContent:  jsontypes.NewNormalizedValue(updatedDashboard),
		Timeframe:         types.StringValue(dashboard.Timeframe),
		Name:              types.StringValue(dashboard.Name),
		SchemaVersion:     types.StringValue(dashboard.SchemaVersion),
		LastUpdated:       types.StringValue(time.Now().Format(time.RFC850)),
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
