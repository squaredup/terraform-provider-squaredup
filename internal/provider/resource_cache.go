package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &CacheResource{}
	_ resource.ResourceWithConfigure   = &CacheResource{}
	_ resource.ResourceWithImportState = &CacheResource{}
)

func SquaredUpCacheResource() resource.Resource {
	return &CacheResource{}
}

type CacheResource struct {
	client *SquaredUpClient
}

type squaredupCache struct {
	ID               types.String         `tfsdk:"id"`
	DisplayName      types.String         `tfsdk:"display_name"`
	DataSourceID     types.String         `tfsdk:"data_source_id"`
	DataStreamID     types.String         `tfsdk:"data_stream_id"`
	DataStreamConfig jsontypes.Normalized `tfsdk:"data_stream_config"`
	SupportedScope   types.String         `tfsdk:"supported_scope"`
	Scope            jsontypes.Normalized `tfsdk:"scope"`
	ObjectIDs        types.Set            `tfsdk:"object_ids"`
	Timeframe        types.String         `tfsdk:"timeframe"`
	FrequencyMinutes types.Int64          `tfsdk:"frequency_minutes"`
	Enabled          types.Bool           `tfsdk:"enabled"`
	TableName        types.String         `tfsdk:"table_name"`
}

func (r *CacheResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache"
}

func (r *CacheResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A cache keeps a full copy of a data stream, so that tiles load fast and the results come back complete, even when the live source is slow or rate limited. " +
			"Data is synced into the cache with a frequency that you choose (from every 5 minutes up to 24 hours), and tiles can then query the cached data instead of hitting the live data stream each time. " +
			"Only raw data is cached, so one cache can be used by multiple tiles in different ways.\n\n" +
			"Caches are available with an Enterprise plan. " +
			"Creating a cache immediately triggers its first sync.\n\n" +
			"Destroying this resource permanently deletes the cache's table and all its data from the warehouse, and any tiles using the cache stop working immediately. Set `enabled = false` to pause syncing instead.\n\n" +
			"See [Cache](https://docs.squaredup.com/features/connect-and-explore/data-streams/cache) for more about how caching works.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the cache",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The cache name. It must be 2 to 24 characters, include at least one letter or number, and be unique among caches of the same data stream. The table name is generated from the cache name when the cache is created, so renaming a cache doesn't change its table name.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 24),
					stringvalidator.RegexMatches(regexp.MustCompile(`[a-zA-Z0-9]`), "must include at least one letter or number"),
				},
			},
			"data_source_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data source the cache is created from. A cache is treated as belonging to this data source. Changing this forces a new cache to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_stream_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data stream to cache",
				Required:            true,
			},
			"data_stream_config": schema.StringAttribute{
				MarkdownDescription: "JSON configuration passed to the data stream (its `dataSourceConfig`), for data streams that take parameters",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"supported_scope": schema.StringAttribute{
				MarkdownDescription: "How the data stream is scoped to objects: `single`, `list` or `none`. Changing this forces a new cache to be created.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("single", "list", "none"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_ids": schema.SetAttribute{
				MarkdownDescription: "The IDs of the objects to cache data for, for example from the `squaredup_nodes` data source. Conflicts with `scope`.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ConflictsWith(path.MatchRoot("scope")),
				},
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "JSON scope selecting the objects to cache data for, for scopes `object_ids` can't express, such as a Gremlin query. " +
					"A collection reference is replaced with the collection's query when the cache is saved, so later changes to the collection don't change the cache. " +
					"Changes made to this scope outside Terraform are not detected. Conflicts with `object_ids`.",
				Optional:   true,
				CustomType: jsontypes.NormalizedType{},
			},
			"timeframe": schema.StringAttribute{
				MarkdownDescription: "The timeframe passed to the source data stream on each sync, for example `last24hours` or `last7days`. Defaults to `last24hours`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("last24hours"),
			},
			"frequency_minutes": schema.Int64Attribute{
				MarkdownDescription: "The sync frequency: how often the cache syncs the data, in minutes. One of `5`, `15`, `30`, `60`, `120`, `180`, `360`, `720` or `1440`, from every 5 minutes up to the default of 24 hours. Sync more often if you need more up-to-date data, but leave enough time for each sync to finish.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1440),
				Validators: []validator.Int64{
					int64validator.OneOf(5, 15, 30, 60, 120, 180, 360, 720, 1440),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the cache syncs at its sync frequency. Disabling stops syncing and keeps the cached data for 48 hours before it is cleaned up. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"table_name": schema.StringAttribute{
				MarkdownDescription: "The name of the table the cache's data is stored in, for use in SQL queries (Parameters > Advanced query, or the Advanced Query data stream on the Cache data source). It is built from the data source, data stream and cache name when the cache is created, and doesn't change if the cache is renamed.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CacheResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CacheResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan squaredupCache
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataStream, diags := cacheDataStream(plan)
	resp.Diagnostics.Append(diags...)
	scope, diags := cacheScope(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cache := CacheCreate{
		DisplayName:      plan.DisplayName.ValueString(),
		PluginConfigID:   plan.DataSourceID.ValueString(),
		DataStream:       dataStream,
		SupportedScope:   plan.SupportedScope.ValueString(),
		Scope:            scope,
		Timeframe:        plan.Timeframe.ValueString(),
		FrequencyMinutes: plan.FrequencyMinutes.ValueInt64Pointer(),
		SyncMode:         "full",
	}

	cacheId, err := r.client.CreateCache(cache)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create cache",
			err.Error(),
		)
		return
	}

	// A new cache is always created enabled
	if !plan.Enabled.ValueBool() {
		err = r.client.SetCacheEnabled(cacheId, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to disable cache",
				err.Error(),
			)
			return
		}
	}

	r.readIntoState(ctx, cacheId, plan, &resp.State, &resp.Diagnostics)
}

func (r *CacheResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state squaredupCache
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cache, err := r.client.GetCache(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get cache",
			err.Error(),
		)
		return
	}

	if cache == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, diags := cacheState(ctx, cache, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *CacheResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state squaredupCache
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch := CachePatch{}
	changed := false

	if !plan.DisplayName.Equal(state.DisplayName) {
		patch.DisplayName = plan.DisplayName.ValueStringPointer()
		changed = true
	}

	if !plan.DataStreamID.Equal(state.DataStreamID) || !plan.DataStreamConfig.Equal(state.DataStreamConfig) {
		dataStream, diags := cacheDataStream(plan)
		resp.Diagnostics.Append(diags...)
		patch.DataStream = &dataStream
		changed = true
	}

	if !plan.Scope.Equal(state.Scope) || !plan.ObjectIDs.Equal(state.ObjectIDs) {
		scope, diags := cacheScope(ctx, plan)
		resp.Diagnostics.Append(diags...)
		if scope == nil {
			// An empty list is how the API is told the scope was cleared
			scope = []interface{}{}
		}
		patch.Scope = scope
		changed = true
	}

	if !plan.Timeframe.Equal(state.Timeframe) {
		patch.Timeframe = plan.Timeframe.ValueStringPointer()
		changed = true
	}

	if !plan.FrequencyMinutes.Equal(state.FrequencyMinutes) {
		patch.FrequencyMinutes = plan.FrequencyMinutes.ValueInt64Pointer()
		changed = true
	}

	if resp.Diagnostics.HasError() {
		return
	}

	cacheId := state.ID.ValueString()

	if changed {
		err := r.client.PatchCache(cacheId, patch)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update cache",
				err.Error(),
			)
			return
		}
	}

	if !plan.Enabled.Equal(state.Enabled) {
		err := r.client.SetCacheEnabled(cacheId, plan.Enabled.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update cache enabled state",
				err.Error(),
			)
			return
		}
	}

	r.readIntoState(ctx, cacheId, plan, &resp.State, &resp.Diagnostics)
}

func (r *CacheResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state squaredupCache
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCache(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete cache",
			err.Error(),
		)
		return
	}
}

func (r *CacheResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// readIntoState fills computed attributes, such as defaults the server applies, after a write.
func (r *CacheResource) readIntoState(ctx context.Context, cacheId string, plan squaredupCache, state stateSetter, diagnostics *diag.Diagnostics) {
	cache, err := r.client.GetCache(cacheId)
	if err == nil && cache == nil {
		err = fmt.Errorf("cache %s was not found after it was written", cacheId)
	}
	if err != nil {
		diagnostics.AddError(
			"Unable to get cache",
			err.Error(),
		)
		return
	}

	newState, diags := cacheState(ctx, cache, plan)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.Set(ctx, &newState)...)
}

type stateSetter interface {
	Set(ctx context.Context, val interface{}) diag.Diagnostics
}

func cacheDataStream(plan squaredupCache) (CacheDataStream, diag.Diagnostics) {
	dataStream := CacheDataStream{ID: plan.DataStreamID.ValueString()}
	if plan.DataStreamConfig.IsNull() {
		return dataStream, nil
	}

	diags := plan.DataStreamConfig.Unmarshal(&dataStream.DataSourceConfig)
	return dataStream, diags
}

// objectIDsBinding follows the SquaredUp UI, which finds the object list in a scope by the
// binding name's ids_ prefix.
const objectIDsBinding = "ids_objectIds"

var objectIDsQueryPattern = regexp.MustCompile(`^g\.V\(\)\.has\('id', within\((ids_\w+)\)\)$`)

// cacheScope returns the scope to send for the plan's object_ids or scope, or nil for neither.
func cacheScope(ctx context.Context, plan squaredupCache) (interface{}, diag.Diagnostics) {
	if !plan.ObjectIDs.IsNull() {
		ids := []string{}
		diags := plan.ObjectIDs.ElementsAs(ctx, &ids, false)
		sort.Strings(ids)

		return map[string]interface{}{
			"query":       "g.V().has('id', within(" + objectIDsBinding + "))",
			"bindings":    map[string]interface{}{objectIDsBinding: ids},
			"queryDetail": map[string]interface{}{"ids": ids},
		}, diags
	}

	if plan.Scope.IsNull() {
		return nil, nil
	}

	var value interface{}
	diags := plan.Scope.Unmarshal(&value)
	return value, diags
}

// scopeObjectIDs returns the object IDs of a stored scope that is a plain list of objects, as
// object_ids and the SquaredUp UI both save one, or nil for any other scope.
func scopeObjectIDs(scope interface{}) []string {
	scopeMap, ok := scope.(map[string]interface{})
	if !ok {
		return nil
	}

	query, _ := scopeMap["query"].(string)
	match := objectIDsQueryPattern.FindStringSubmatch(query)
	bindings, _ := scopeMap["bindings"].(map[string]interface{})
	if match == nil || bindings == nil {
		return nil
	}

	values, ok := bindings[match[1]].([]interface{})
	if !ok || len(values) == 0 {
		return nil
	}

	ids := make([]string, 0, len(values))
	for _, value := range values {
		id, ok := value.(string)
		if !ok {
			return nil
		}
		ids = append(ids, id)
	}

	return ids
}

// cacheState maps a cache read from the API onto Terraform state. The scope is kept from prior,
// because the server rewrites collection references into their queries on save. Object IDs are
// read back unless the configuration uses scope, so a list changed outside Terraform shows as drift.
func cacheState(ctx context.Context, cache *CacheRead, prior squaredupCache) (squaredupCache, diag.Diagnostics) {
	var diags diag.Diagnostics

	displayName := cache.DisplayName
	if displayName == "" {
		displayName = cache.Name
	}

	state := squaredupCache{
		ID:               types.StringValue(cache.ID),
		DisplayName:      types.StringValue(displayName),
		Scope:            prior.Scope,
		FrequencyMinutes: types.Int64Value(cache.FrequencyMinutes),
		Enabled:          types.BoolValue(cache.Enabled),
		DataStreamConfig: prior.DataStreamConfig,
		Timeframe:        prior.Timeframe,
	}

	if len(cache.Steps) == 0 {
		diags.AddError("Unexpected cache response", "The cache "+cache.ID+" has no steps.")
		return state, diags
	}
	step := cache.Steps[0]
	mapping := step.CacheMapping

	state.DataSourceID = types.StringValue(step.DataStream.PluginConfigID)
	state.DataStreamID = types.StringValue(step.DataStream.ID)
	state.SupportedScope = types.StringValue(step.SupportedScope)
	state.TableName = types.StringValue(mapping.TableName)

	if mapping.SyncMode != "" && mapping.SyncMode != "full" {
		diags.AddWarning("Unsupported cache sync mode",
			"The cache "+cache.ID+" uses \""+mapping.SyncMode+"\" sync, which this resource does not manage. Caches created by this resource always use full sync.")
	}

	// Older caches may have no frequency stored, which the SquaredUp UI shows as 24 hours
	if cache.FrequencyMinutes == 0 {
		state.FrequencyMinutes = types.Int64Value(1440)
	}

	// Timeframes set outside Terraform can be objects, which this resource does not model
	if timeframe, ok := step.Timeframe.(string); ok {
		state.Timeframe = types.StringValue(timeframe)
	}

	if len(step.DataStream.DataSourceConfig) > 0 {
		dataSourceConfig, err := json.Marshal(step.DataStream.DataSourceConfig)
		if err != nil {
			diags.AddError("Unable to read cache data stream config", err.Error())
			return state, diags
		}
		state.DataStreamConfig = jsontypes.NewNormalizedValue(string(dataSourceConfig))
	} else if !prior.DataStreamConfig.IsNull() {
		state.DataStreamConfig = jsontypes.NewNormalizedValue("{}")
	}

	state.ObjectIDs = types.SetNull(types.StringType)
	if objectIDs := scopeObjectIDs(step.Scope); objectIDs != nil && prior.Scope.IsNull() {
		ids, setDiags := types.SetValueFrom(ctx, types.StringType, objectIDs)
		diags.Append(setDiags...)
		state.ObjectIDs = ids
	}

	return state, diags
}
