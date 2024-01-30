package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = &workspaceResource{}
	_ resource.ResourceWithConfigure   = &workspaceResource{}
	_ resource.ResourceWithImportState = &workspaceResource{}
)

func SquaredupWorkspaceResource() resource.Resource {
	return &workspaceResource{}
}

type workspaceResource struct {
	client *SquaredUpClient
}

type workspace struct {
	DisplayName             types.String   `tfsdk:"display_name"`
	Description             types.String   `tfsdk:"description"`
	Type                    types.String   `tfsdk:"type"`
	Tags                    []types.String `tfsdk:"tags"`
	DataSourcesLinks        []types.String `tfsdk:"datasources_links"`
	WorkspacesLinks         []types.String `tfsdk:"workspaces_links"`
	DashboardSharingEnabled types.Bool     `tfsdk:"allow_dashboard_sharing"`
	ID                      types.String   `tfsdk:"id"`
	LastUpdated             types.String   `tfsdk:"last_updated"`
}

func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Each workspace has its own dashboards, data sources, monitors and scopes.",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				Description: "The display name of the workspace (Displayed in SquaredUp)",
				Required:    true,
			},

			"description": schema.StringAttribute{
				Description: "The description of the workspace",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Workspace type that can be one of: 'service', 'team', 'application', 'platform', 'product', 'business service', 'microservice', 'customer', 'website', 'component', 'resource', 'system', 'folder', 'other'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{stringvalidator.OneOf(
					"service",
					"team",
					"application",
					"platform",
					"product",
					"business service",
					"microservice",
					"customer",
					"website",
					"component",
					"resource",
					"system",
					"folder",
					"other",
				)},
			},
			"tags": schema.ListAttribute{
				Description: "The tags of the workspace",
				Optional:    true,
				Computed:    true,
				ElementType: basetypes.StringType{},
				Default:     listdefault.StaticValue(basetypes.ListValue{}),
			},
			"datasources_links": schema.ListAttribute{
				Description: "Links to plugins",
				Optional:    true,
				Computed:    true,
				ElementType: basetypes.StringType{},
				Default:     listdefault.StaticValue(basetypes.ListValue{}),
			},
			"workspaces_links": schema.ListAttribute{
				Description: "Links to workspaces",
				Optional:    true,
				Computed:    true,
				ElementType: basetypes.StringType{},
				Default:     listdefault.StaticValue(basetypes.ListValue{}),
			},
			"allow_dashboard_sharing": schema.BoolAttribute{
				Description: "Allow dashboards in this workspace to be shared with anyone",
				Optional:    true,
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the workspace",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "The last time the workspace was updated",
				Computed:    true,
			},
		},
	}
}

func (r *workspaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workspace
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	workspacePayload := map[string]interface{}{
		"displayName": plan.DisplayName.ValueString(),
		"links": map[string]interface{}{
			"plugins":    SafeStringConversion(plan.DataSourcesLinks),
			"workspaces": SafeStringConversion(plan.WorkspacesLinks),
		},
		"linkToWorkspaces": true,
		"properties": map[string]interface{}{
			"DashboardSharingEnabled": plan.DashboardSharingEnabled.ValueBool(),
			"tags":                    SafeStringConversion(plan.Tags),
			"description":             plan.Description.ValueString(),
		},
	}

	if properties, ok := workspacePayload["properties"].(map[string]interface{}); ok {
		if plan.Type.ValueString() != "" {
			properties["type"] = plan.Type.ValueString()
		}
	}

	newWorkspace, err := r.client.CreateWorkspace(workspacePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to create workspace",
			err.Error(),
		)
		return
	}

	workspaceID := strings.Trim(newWorkspace, `"`)

	readWorkspace, err := r.client.GetWorkspace(workspaceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to get workspace",
			err.Error(),
		)
		return
	}

	workspace := workspace{
		DisplayName:             types.StringValue(readWorkspace.DisplayName),
		ID:                      types.StringValue(readWorkspace.ID),
		Description:             types.StringValue(readWorkspace.Data.Properties.Description),
		Type:                    types.StringValue(readWorkspace.Data.Properties.Type),
		DashboardSharingEnabled: types.BoolValue(readWorkspace.Data.Properties.DashboardSharingEnabled),
		LastUpdated:             types.StringValue(time.Now().Format(time.RFC850)),
	}

	if len(readWorkspace.Data.Properties.Tags) > 0 {
		workspace.Tags = toBasetypesStringSlice(readWorkspace.Data.Properties.Tags)
	}

	if len(readWorkspace.Data.Links.Plugins) > 0 {
		workspace.DataSourcesLinks = toBasetypesStringSlice(readWorkspace.Data.Links.Plugins)
	}

	if len(readWorkspace.Data.Links.Workspaces) > 0 {
		workspace.WorkspacesLinks = toBasetypesStringSlice(readWorkspace.Data.Links.Workspaces)
	}

	diags = resp.State.Set(ctx, workspace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspace
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	readWorkspace, err := r.client.GetWorkspace(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to get workspace",
			err.Error(),
		)
		return
	}

	workspace := workspace{
		DisplayName:             types.StringValue(readWorkspace.DisplayName),
		ID:                      types.StringValue(readWorkspace.ID),
		Description:             types.StringValue(readWorkspace.Data.Properties.Description),
		Type:                    types.StringValue(readWorkspace.Data.Properties.Type),
		DashboardSharingEnabled: types.BoolValue(readWorkspace.Data.Properties.DashboardSharingEnabled),
	}

	if len(readWorkspace.Data.Properties.Tags) > 0 {
		workspace.Tags = toBasetypesStringSlice(readWorkspace.Data.Properties.Tags)
	}

	if len(readWorkspace.Data.Links.Plugins) > 0 {
		workspace.DataSourcesLinks = toBasetypesStringSlice(readWorkspace.Data.Links.Plugins)
	}

	if len(readWorkspace.Data.Links.Workspaces) > 0 {
		workspace.WorkspacesLinks = toBasetypesStringSlice(readWorkspace.Data.Links.Workspaces)
	}

	diags = resp.State.Set(ctx, workspace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workspace
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	var state workspace
	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	workspacePayload := map[string]interface{}{
		"displayName": plan.DisplayName.ValueString(),
		"links": map[string]interface{}{
			"plugins":    SafeStringConversion(plan.DataSourcesLinks),
			"workspaces": SafeStringConversion(plan.WorkspacesLinks),
		},
		"linkToWorkspaces": true,
		"properties": map[string]interface{}{
			"DashboardSharingEnabled": plan.DashboardSharingEnabled.ValueBool(),
			"tags":                    SafeStringConversion(plan.Tags),
			"description":             plan.Description.ValueString(),
		},
	}

	if properties, ok := workspacePayload["properties"].(map[string]interface{}); ok {
		if plan.Type.ValueString() != "" {
			properties["type"] = plan.Type.ValueString()
		}
	}

	err := r.client.UpdateWorkspace(state.ID.ValueString(), workspacePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to update workspace",
			err.Error(),
		)
		return
	}

	readWorkspace, err := r.client.GetWorkspace(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to get workspace",
			err.Error(),
		)
		return
	}

	workspace := workspace{
		DisplayName:             types.StringValue(readWorkspace.DisplayName),
		ID:                      types.StringValue(readWorkspace.ID),
		Description:             types.StringValue(readWorkspace.Data.Properties.Description),
		Type:                    types.StringValue(readWorkspace.Data.Properties.Type),
		DashboardSharingEnabled: types.BoolValue(readWorkspace.Data.Properties.DashboardSharingEnabled),
		LastUpdated:             types.StringValue(time.Now().Format(time.RFC850)),
	}

	if len(readWorkspace.Data.Properties.Tags) > 0 {
		workspace.Tags = toBasetypesStringSlice(readWorkspace.Data.Properties.Tags)
	}

	if len(readWorkspace.Data.Links.Plugins) > 0 {
		workspace.DataSourcesLinks = toBasetypesStringSlice(readWorkspace.Data.Links.Plugins)
	}

	if len(readWorkspace.Data.Links.Workspaces) > 0 {
		workspace.WorkspacesLinks = toBasetypesStringSlice(readWorkspace.Data.Links.Workspaces)
	}

	diags = resp.State.Set(ctx, workspace)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workspace
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteWorkspace(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to delete workspace",
			err.Error(),
		)
		return
	}
}

func (r *workspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func toBasetypesStringSlice(input []string) []basetypes.StringValue {
	result := make([]basetypes.StringValue, len(input))
	for i, s := range input {
		result[i] = basetypes.NewStringValue(s)
	}

	return result
}

func SafeStringConversion(inputList []basetypes.StringValue) []string {
	result := make([]string, len(inputList))
	for i, item := range inputList {
		result[i] = item.ValueString()
	}

	return result
}
