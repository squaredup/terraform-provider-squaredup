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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pborman/uuid"
)

var (
	_ resource.Resource                = &ScopeResource{}
	_ resource.ResourceWithConfigure   = &ScopeResource{}
	_ resource.ResourceWithImportState = &ScopeResource{}
)

func SquaredUpScopeResource() resource.Resource {
	return &ScopeResource{}
}

type ScopeResource struct {
	client *SquaredUpClient
}

type SquaredUpScope struct {
	ScopeID       types.String   `tfsdk:"id"`
	DisplayName   types.String   `tfsdk:"display_name"`
	ScopeType     types.String   `tfsdk:"scope_type"`
	LastUpdated   types.String   `tfsdk:"last_updated"`
	WorkspaceId   types.String   `tfsdk:"workspace_id"`
	DataSourceId  []types.String `tfsdk:"data_source_id"`
	Types         []types.String `tfsdk:"types"`
	SearchQuery   types.String   `tfsdk:"search_query"`
	AdvancedQuery types.String   `tfsdk:"advanced_query"`
	NodeIds       []types.String `tfsdk:"node_ids"`
	Query         types.String   `tfsdk:"query"`
}

func (r *ScopeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scope"
}

func (r *ScopeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A collection (previously known as scope) contains objects indexed by data sources. A collection can be used as a filter when configuring dashboards and tiles.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the scope",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name for the scope",
				Required:            true,
			},
			"scope_type": schema.StringAttribute{
				MarkdownDescription: "Type of the scope. Either 'dynamic' or 'fixed'",
				Required:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					"dynamic",
					"fixed",
					"advanced",
				)},
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "Last updated timestamp",
				Computed:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace",
				Required:            true,
			},
			"data_source_id": schema.ListAttribute{
				MarkdownDescription: "IDs of the data sources to filter the scope",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"types": schema.ListAttribute{
				MarkdownDescription: "Node types to filter the scope",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"search_query": schema.StringAttribute{
				MarkdownDescription: "Search query",
				Optional:            true,
			},
			"advanced_query": schema.StringAttribute{
				MarkdownDescription: "Advanced query (Gremlin)",
				Optional:            true,
			},
			"node_ids": schema.ListAttribute{
				MarkdownDescription: "IDs of the nodes that scope will contain",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Query for the scope",
				Computed:            true,
			},
		},
	}
}

func (r *ScopeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Scope Configure Type",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ScopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := SquaredUpScope{}
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	var scopePayload ScopeCreate
	var err error

	switch plan.ScopeType.ValueString() {
	case "fixed":
		scopePayload, err = buildFixedScope(plan)
	case "dynamic":
		scopePayload, err = buildDynamicScope(plan)
	case "advanced":
		scopePayload, err = buildAdvancedScope(plan)
		resp.Diagnostics.AddWarning("You are using advanced query, the UI may not be able to render the scope correctly!", "")
	default:
		err = fmt.Errorf("invalid scope type: %s", plan.ScopeType.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to build scope payload", err.Error())
		return
	}

	scopeID, err := r.client.CreateScope(scopePayload, plan.WorkspaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create scope", err.Error())
		return
	}

	readScope, err := r.client.GetScope(scopeID, plan.WorkspaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read scope", err.Error())
		return
	}

	state := createState(plan, readScope)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ScopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SquaredUpScope
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	readScope, err := r.client.GetScope(state.ScopeID.ValueString(), state.WorkspaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read scope", err.Error())
		return
	}

	state = createState(state, readScope)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ScopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := SquaredUpScope{}
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	var scopePayload ScopeCreate
	var err error

	switch plan.ScopeType.ValueString() {
	case "fixed":
		scopePayload, err = buildFixedScope(plan)
	case "dynamic":
		scopePayload, err = buildDynamicScope(plan)
	case "advanced":
		scopePayload, err = buildAdvancedScope(plan)
	default:
		err = fmt.Errorf("invalid scope type: %s", plan.ScopeType.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to build scope payload", err.Error())
		return
	}

	err = r.client.UpdateScope(plan.ScopeID.ValueString(), scopePayload, plan.WorkspaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update scope", err.Error())
		return
	}

	readScope, err := r.client.GetScope(plan.ScopeID.ValueString(), plan.WorkspaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read scope", err.Error())
		return
	}

	state := createState(plan, readScope)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ScopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SquaredUpScope
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteScope(state.ScopeID.ValueString(), state.WorkspaceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete scope", err.Error())
		return
	}
}

func (r *ScopeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: workspace_id,scope_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func generateRandomID() string {
	id := uuid.NewRandom().String()
	idString := strings.ReplaceAll(id, "-", "")
	randomId := idString[:20]
	return randomId
}

func buildFixedScope(plan SquaredUpScope) (ScopeCreate, error) {
	var scopePayload ScopeCreate

	if len(plan.NodeIds) == 0 {
		return scopePayload, fmt.Errorf("node_ids is required for fixed scope")
	}

	if len(plan.DataSourceId) != 0 || len(plan.Types) != 0 || plan.SearchQuery.ValueString() != "" || plan.AdvancedQuery.ValueString() != "" {
		return scopePayload, fmt.Errorf("data_source_id, types, search_query and advanced_query are not allowed for fixed scope")
	}

	query := "g.V().hasId(within("
	for i, id := range plan.NodeIds {
		if i != 0 {
			query += ","
		}
		query += fmt.Sprintf("'%s'", id.ValueString())
	}
	query += "))"

	nodeIds := make([]string, 0)
	for _, id := range plan.NodeIds {
		nodeIds = append(nodeIds, id.ValueString())
	}

	scopePayload = ScopeCreate{
		Scope: Scope{
			Name:    plan.DisplayName.ValueString(),
			Version: 2,
			Query:   query,
			QueryDetail: ScopeQueryDetail{
				IDs: nodeIds,
			},
		},
	}

	return scopePayload, nil
}

func buildDynamicScope(plan SquaredUpScope) (ScopeCreate, error) {
	var scopePayload ScopeCreate

	if len(plan.NodeIds) != 0 {
		return scopePayload, fmt.Errorf("node_ids are not allowed for dynamic scope")
	}

	if plan.SearchQuery.ValueString() == "" {
		return scopePayload, fmt.Errorf("search_query is required for dynamic scope")
	}

	var query string
	booleanQuery := plan.SearchQuery.ValueString()

	bindings := make(map[string]interface{})
	queryDetail := ScopeQueryDetail{
		BooleanQuery: booleanQuery,
	}

	booleanQueryBinding := "booleanQuery_" + generateRandomID()
	bindings[booleanQueryBinding] = booleanQuery
	query = fmt.Sprintf("g.V().has('__search', __matchesQuery(%s))", booleanQueryBinding)

	datasourceQueryBinding := "plugins_" + generateRandomID()
	if len(plan.DataSourceId) != 0 {
		var dataSourceBindingIDs []string
		for _, id := range plan.DataSourceId {
			dataSourceBindingIDs = append(dataSourceBindingIDs, id.ValueString())
		}
		bindings[datasourceQueryBinding] = dataSourceBindingIDs
		query += fmt.Sprintf(".or(__.has('__configId', within(%s)))", datasourceQueryBinding)
		plugins := make([]ScopeQueryDetailPlugin, 0)
		for _, id := range plan.DataSourceId {
			plugins = append(plugins, ScopeQueryDetailPlugin{Value: id.ValueString()})
		}
		queryDetail.Plugins = plugins
	}

	typesQueryBinding := "types_" + generateRandomID()
	if len(plan.Types) != 0 {
		var typesBindingIDs []string
		for _, t := range plan.Types {
			typesBindingIDs = append(typesBindingIDs, t.ValueString())
		}
		bindings[typesQueryBinding] = typesBindingIDs
		query += fmt.Sprintf(".has('type', within(%s))", typesQueryBinding)
		types := make([]ScopeQueryDetailType, 0)
		for _, t := range plan.Types {
			types = append(types, ScopeQueryDetailType{Value: t.ValueString()})
		}
		queryDetail.Types = types
	}

	query += ".order().by('__name').hasNot('__canonicalType').limit(500)"

	scopePayload = ScopeCreate{
		Scope: Scope{
			Name:        plan.DisplayName.ValueString(),
			Version:     2,
			Query:       query,
			Bindings:    bindings,
			QueryDetail: queryDetail,
		},
	}

	return scopePayload, nil
}

func buildAdvancedScope(plan SquaredUpScope) (ScopeCreate, error) {
	var scopePayload ScopeCreate

	if plan.AdvancedQuery.ValueString() == "" {
		return scopePayload, fmt.Errorf("advanced_query is required for advanced scope")
	}

	if len(plan.NodeIds) != 0 || len(plan.DataSourceId) != 0 || len(plan.Types) != 0 || plan.SearchQuery.ValueString() != "" {
		return scopePayload, fmt.Errorf("node_ids, data_source_id, types and search_query are not supported for advanced scope")
	}

	query := plan.AdvancedQuery.ValueString()

	scopePayload = ScopeCreate{
		Scope: Scope{
			Name:    plan.DisplayName.ValueString(),
			Version: 2,
			Query:   query,
		},
	}

	return scopePayload, nil

}

func createState(plan SquaredUpScope, readScope *ScopeRead) SquaredUpScope {
	state := SquaredUpScope{
		ScopeID:     types.StringValue(readScope.ID),
		DisplayName: types.StringValue(readScope.DisplayName),
		ScopeType:   types.StringValue(plan.ScopeType.ValueString()),
		LastUpdated: types.StringValue(time.Now().Format(time.RFC850)),
		WorkspaceId: types.StringValue(readScope.WorkspaceID),
		Query:       types.StringValue(readScope.Data.Query),
	}

	switch plan.ScopeType.ValueString() {
	case "fixed":
		if len(plan.NodeIds) != 0 {
			state.NodeIds = plan.NodeIds
		}
	case "dynamic":
		if len(plan.DataSourceId) != 0 {
			state.DataSourceId = plan.DataSourceId
		}
		if len(plan.Types) != 0 {
			state.Types = plan.Types
		}
		if plan.SearchQuery.ValueString() != "" {
			state.SearchQuery = plan.SearchQuery
		}
	case "advanced":
		state.AdvancedQuery = plan.AdvancedQuery
	}

	return state
}
