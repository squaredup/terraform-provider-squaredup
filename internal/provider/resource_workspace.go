package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	DisplayName             types.String `tfsdk:"display_name"`
	Description             types.String `tfsdk:"description"`
	Type                    types.String `tfsdk:"type"`
	Tags                    types.List   `tfsdk:"tags"`
	DataSourcesLinks        types.List   `tfsdk:"datasources_links"`
	WorkspacesLinks         types.List   `tfsdk:"workspaces_links"`
	ReadWorkspacesLinks     types.List   `tfsdk:"read_workspaces_links"`
	DashboardSharingEnabled types.Bool   `tfsdk:"allow_dashboard_sharing"`
	ID                      types.String `tfsdk:"id"`
	LastUpdated             types.String `tfsdk:"last_updated"`
}

func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (r *workspaceResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Each workspace has its own dashboards, data sources, monitors and scopes",
		Attributes: map[string]schema.Attribute{
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name for the workspace",
				Required:            true,
			},

			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the workspace",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Workspace type that can be one of: 'service', 'team', 'application', 'platform', 'product', 'business service', 'microservice', 'customer', 'website', 'component', 'resource', 'system', 'folder', 'other'.",
				Optional:            true,
				Computed:            true,
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
				Default: stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags for the workspace",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"datasources_links": schema.ListAttribute{
				MarkdownDescription: "IDs of Data Sources to link to this workspace",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"workspaces_links": schema.ListAttribute{
				MarkdownDescription: "IDs of Workspaces to link to this workspace",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"read_workspaces_links": schema.ListAttribute{
				MarkdownDescription: "IDs of Workspaces linked to this workspace",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"allow_dashboard_sharing": schema.BoolAttribute{
				MarkdownDescription: "Allow dashboards in this workspace to be shared",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the workspace",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The last time the workspace was updated",
				Computed:            true,
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
			"Unexpected Workspace Configure Type",
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

	workspacePayload := GenerateWorkspacePayload(plan)

	workspaceID, err := r.client.CreateWorkspace(workspacePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to create workspace",
			err.Error(),
		)
		return
	}

	readWorkspace, err := r.client.GetWorkspace(workspaceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to get workspace",
			err.Error(),
		)
		return
	}

	workspace := GenerateWorkspaceState(readWorkspace)
	workspace.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	workspace.WorkspacesLinks = plan.WorkspacesLinks

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

	workspace := GenerateWorkspaceState(readWorkspace)
	workspace.WorkspacesLinks = state.WorkspacesLinks

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

	workspacePayload := GenerateWorkspacePayload(plan)

	err := r.client.UpdateWorkspace(plan.ID.ValueString(), workspacePayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to update workspace",
			err.Error(),
		)
		return
	}

	readWorkspace, err := r.client.GetWorkspace(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to get workspace",
			err.Error(),
		)
		return
	}

	workspace := GenerateWorkspaceState(readWorkspace)
	workspace.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	workspace.WorkspacesLinks = plan.WorkspacesLinks

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

func GenerateWorkspacePayload(plan workspace) map[string]interface{} {

	var linkedPlugins []string = []string{}
	if !plan.DataSourcesLinks.IsNull() && !plan.DataSourcesLinks.IsUnknown() {
		var plugins []types.String
		plan.DataSourcesLinks.ElementsAs(context.TODO(), &plugins, false)
		for _, plugin := range plugins {
			linkedPlugins = append(linkedPlugins, plugin.ValueString())
		}
	}
	var linkedWorkspaces []string = []string{}
	if !plan.WorkspacesLinks.IsNull() && !plan.WorkspacesLinks.IsUnknown() {
		var workspaces []types.String
		plan.WorkspacesLinks.ElementsAs(context.TODO(), &workspaces, false)
		for _, workspace := range workspaces {
			linkedWorkspaces = append(linkedWorkspaces, workspace.ValueString())
		}
	}

	var tags []string = []string{}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tagList []types.String
		plan.Tags.ElementsAs(context.TODO(), &tagList, false)
		for _, tag := range tagList {
			tags = append(tags, tag.ValueString())
		}
	}

	workspacePayload := map[string]interface{}{
		"displayName":      plan.DisplayName.ValueString(),
		"linkToWorkspaces": linkedWorkspaces == nil,
		"links": map[string]interface{}{
			"plugins":    linkedPlugins,
			"workspaces": linkedWorkspaces,
		},
		"properties": map[string]interface{}{
			"openAccessEnabled": plan.DashboardSharingEnabled.ValueBool(),
			"tags":              tags,
			"description":       plan.Description.ValueString(),
		},
	}

	// SquaredUp API doesn't allow empty string for type, so only add it if it's not empty
	if properties, ok := workspacePayload["properties"].(map[string]interface{}); ok {
		if plan.Type.ValueString() != "" {
			properties["type"] = plan.Type.ValueString()
		}
	}

	return workspacePayload
}

func GenerateWorkspaceState(workspaceRead *WorkspaceRead) workspace {
	workspace := workspace{
		DisplayName:             types.StringValue(workspaceRead.DisplayName),
		ID:                      types.StringValue(workspaceRead.ID),
		Description:             types.StringValue(workspaceRead.Data.Properties.Description),
		Type:                    types.StringValue(workspaceRead.Data.Properties.Type),
		DashboardSharingEnabled: types.BoolValue(workspaceRead.Data.Properties.DashboardSharingEnabled),
		Tags:                    types.ListNull(types.StringType),
		DataSourcesLinks:        types.ListNull(types.StringType),
		WorkspacesLinks:         types.ListNull(types.StringType),
		ReadWorkspacesLinks:     types.ListNull(types.StringType),
	}

	if len(workspaceRead.Data.Properties.Tags) > 0 {
		tags := []attr.Value{}
		for _, tag := range workspaceRead.Data.Properties.Tags {
			tags = append(tags, types.StringValue(tag))
		}
		workspace.Tags = types.ListValueMust(types.StringType, tags)
	}

	if len(workspaceRead.Data.Links.Plugins) > 0 {
		plugins := []attr.Value{}
		for _, plugin := range workspaceRead.Data.Links.Plugins {
			plugins = append(plugins, types.StringValue(plugin))
		}
		workspace.DataSourcesLinks = types.ListValueMust(types.StringType, plugins)
	}

	if len(workspaceRead.Data.Links.Workspaces) > 0 {
		workspaces := []attr.Value{}
		for _, workspace := range workspaceRead.Data.Links.Workspaces {
			workspaces = append(workspaces, types.StringValue(workspace))
		}
		workspace.ReadWorkspacesLinks = types.ListValueMust(types.StringType, workspaces)
	}

	return workspace
}
