package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/pborman/uuid"
)

// suffix keeps the data source name unique, as CI runs these tests in parallel against one tenant.
func cacheTestConfig(suffix string, displayName string, frequencyMinutes string, enabled string, objectIDs string) string {
	return providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - Cache Test - ` + suffix + `"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

data "squaredup_data_streams" "database_free_space" {
	data_source_id              = data.squaredup_datasources.sample_data.plugins[0].id
	data_stream_definition_name = "database-free-space"
}

data "squaredup_nodes" "database_1" {
	depends_on     = [squaredup_datasource.sample_data_source]
	data_source_id = squaredup_datasource.sample_data_source.id
	node_source_id = "sample-database-1"
}

data "squaredup_nodes" "database_2" {
	depends_on     = [squaredup_datasource.sample_data_source]
	data_source_id = squaredup_datasource.sample_data_source.id
	node_source_id = "sample-database-2"
}

resource "squaredup_cache" "test" {
	display_name      = "` + displayName + `"
	data_source_id    = squaredup_datasource.sample_data_source.id
	data_stream_id    = data.squaredup_data_streams.database_free_space.data_streams[0].id
	object_ids        = ` + objectIDs + `
	frequency_minutes = ` + frequencyMinutes + `
	enabled           = ` + enabled + `
}
`
}

const (
	bothDatabases = "[data.squaredup_nodes.database_1.node_properties[0].id, data.squaredup_nodes.database_2.node_properties[0].id]"
	oneDatabase   = "[data.squaredup_nodes.database_1.node_properties[0].id]"
)

func TestAccResourceCache(t *testing.T) {
	// Cache names are capped at 24 characters
	suffix := uuid.NewRandom().String()[:8]

	// An edit must change the existing cache, keeping its ID and table, rather than replace it
	sameID := statecheck.CompareValue(compare.ValuesSame())
	sameTable := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: cacheTestConfig(suffix, "TF Cache "+suffix, "60", "true", bothDatabases),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_cache.test", "display_name", "TF Cache "+suffix),
					resource.TestCheckResourceAttr("squaredup_cache.test", "frequency_minutes", "60"),
					resource.TestCheckResourceAttr("squaredup_cache.test", "timeframe", "last24hours"),
					resource.TestCheckResourceAttr("squaredup_cache.test", "enabled", "true"),
					resource.TestCheckResourceAttr("squaredup_cache.test", "object_ids.#", "2"),
					resource.TestCheckResourceAttr("squaredup_cache.test", "supported_scope", "list"),
					resource.TestCheckResourceAttrSet("squaredup_cache.test", "id"),
					resource.TestCheckResourceAttrSet("squaredup_cache.test", "table_name"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					sameID.AddStateValue("squaredup_cache.test", tfjsonpath.New("id")),
					sameTable.AddStateValue("squaredup_cache.test", tfjsonpath.New("table_name")),
				},
			},
			// Import Test
			{
				ResourceName:      "squaredup_cache.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update Test
			{
				Config: cacheTestConfig(suffix, "TF Renamed "+suffix, "120", "false", oneDatabase),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("squaredup_cache.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_cache.test", "display_name", "TF Renamed "+suffix),
					resource.TestCheckResourceAttr("squaredup_cache.test", "frequency_minutes", "120"),
					resource.TestCheckResourceAttr("squaredup_cache.test", "enabled", "false"),
					resource.TestCheckResourceAttr("squaredup_cache.test", "object_ids.#", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					sameID.AddStateValue("squaredup_cache.test", tfjsonpath.New("id")),
					sameTable.AddStateValue("squaredup_cache.test", tfjsonpath.New("table_name")),
				},
			},
		},
	})
}

func TestCacheStateMapsReadResponse(t *testing.T) {
	read := &CacheRead{
		ID:               "smartcachedefinition-abc",
		Name:             "Costs",
		FrequencyMinutes: 60,
		Enabled:          true,
		Steps: []CacheStep{{
			DataStream: CacheDataStream{
				ID:               "datastream-abc",
				PluginConfigID:   "config-abc",
				DataSourceConfig: map[string]interface{}{"region": "eu"},
			},
			Timeframe:      "last7days",
			SupportedScope: "none",
		}},
	}
	read.Steps[0].CacheMapping.TableName = "costs_table"
	read.Steps[0].CacheMapping.SyncMode = "full"

	prior := squaredupCache{Scope: jsontypes.NewNormalizedValue(`{"query":"g.V()"}`)}

	state, diags := cacheState(context.Background(), read, prior)
	if diags.HasError() || diags.WarningsCount() > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.DisplayName.ValueString() != "Costs" {
		t.Errorf("display_name = %q, want the name as the fallback", state.DisplayName.ValueString())
	}
	if state.DataSourceID.ValueString() != "config-abc" || state.DataStreamID.ValueString() != "datastream-abc" {
		t.Errorf("data source or stream not mapped: %v, %v", state.DataSourceID, state.DataStreamID)
	}
	if state.DataStreamConfig.ValueString() != `{"region":"eu"}` {
		t.Errorf("data_stream_config = %s", state.DataStreamConfig.ValueString())
	}
	if !state.Scope.Equal(prior.Scope) {
		t.Errorf("scope should be kept from prior state, got %s", state.Scope.ValueString())
	}
	if state.Timeframe.ValueString() != "last7days" || state.TableName.ValueString() != "costs_table" {
		t.Errorf("timeframe or table_name not mapped: %v, %v", state.Timeframe, state.TableName)
	}
}

func TestCacheStateDefaultsMissingFrequencyToDaily(t *testing.T) {
	read := &CacheRead{ID: "smartcachedefinition-abc", DisplayName: "Costs", Steps: []CacheStep{{}}}

	state, diags := cacheState(context.Background(), read, squaredupCache{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if state.FrequencyMinutes.ValueInt64() != 1440 {
		t.Errorf("a missing frequency should read as 24 hours, got %d", state.FrequencyMinutes.ValueInt64())
	}
}

func TestCacheStateWarnsWhenSyncIsNotFull(t *testing.T) {
	read := &CacheRead{ID: "smartcachedefinition-abc", DisplayName: "Costs", Steps: []CacheStep{{}}}
	read.Steps[0].CacheMapping.SyncMode = "incremental"

	_, diags := cacheState(context.Background(), read, squaredupCache{})
	if diags.HasError() || diags.WarningsCount() != 1 {
		t.Errorf("a cache that is not a full sync should warn without failing, got %v", diags)
	}
}

func TestCacheScopeBuildsObjectIDsScope(t *testing.T) {
	objectIDs, _ := types.SetValueFrom(context.Background(), types.StringType, []string{"node-b", "node-a"})
	plan := squaredupCache{ObjectIDs: objectIDs, Scope: jsontypes.NewNormalizedNull()}

	scope, diags := cacheScope(context.Background(), plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	got, _ := json.Marshal(scope)
	want := `{"bindings":{"ids_objectIds":["node-a","node-b"]},"query":"g.V().has('id', within(ids_objectIds))","queryDetail":{"ids":["node-a","node-b"]}}`
	if string(got) != want {
		t.Errorf("scope = %s, want %s", got, want)
	}
	var returned interface{}
	_ = json.Unmarshal(got, &returned)
	if ids := scopeObjectIDs(returned); len(ids) != 2 || ids[0] != "node-a" {
		t.Errorf("the scope sent should read back as the same objects, got %v", ids)
	}
}

func TestScopeObjectIDsReadsObjectListScopes(t *testing.T) {
	var uiScope interface{}
	_ = json.Unmarshal([]byte(`{"query":"g.V().has('id', within(ids_ncTbNH15BGQ1r6sHKne9))","bindings":{"ids_ncTbNH15BGQ1r6sHKne9":["node-a"]},"queryDetail":{"ids":["node-a"]}}`), &uiScope)
	if ids := scopeObjectIDs(uiScope); len(ids) != 1 || ids[0] != "node-a" {
		t.Errorf("an object list saved by the UI should be read, got %v", ids)
	}

	var queryScope interface{}
	_ = json.Unmarshal([]byte(`{"query":"g.V().has('type', 'sample-database')","bindings":{}}`), &queryScope)
	if ids := scopeObjectIDs(queryScope); ids != nil {
		t.Errorf("a Gremlin query scope is not an object list, got %v", ids)
	}

	if ids := scopeObjectIDs([]interface{}{}); ids != nil {
		t.Errorf("a cleared scope has no objects, got %v", ids)
	}
}

func TestCacheStateReadsObjectIDsUnlessScopeIsConfigured(t *testing.T) {
	var stored interface{}
	_ = json.Unmarshal([]byte(`{"query":"g.V().has('id', within(ids_objectIds))","bindings":{"ids_objectIds":["node-a","node-c"]}}`), &stored)
	read := &CacheRead{ID: "smartcachedefinition-abc", DisplayName: "Costs", Steps: []CacheStep{{Scope: stored}}}

	state, _ := cacheState(context.Background(), read, squaredupCache{Scope: jsontypes.NewNormalizedNull()})
	if len(state.ObjectIDs.Elements()) != 2 {
		t.Errorf("object_ids should be read back so changes made outside Terraform show as drift, got %v", state.ObjectIDs)
	}

	withScope := squaredupCache{Scope: jsontypes.NewNormalizedValue(`{"query":"g.V().has('id', within(ids_objectIds))"}`)}
	state, _ = cacheState(context.Background(), read, withScope)
	if !state.ObjectIDs.IsNull() {
		t.Errorf("object_ids should stay null when the configuration uses scope, got %v", state.ObjectIDs)
	}
}

func TestFindSupportedScopeMatchesTheDataStreamsFunction(t *testing.T) {
	plugin := pluginDataSources{}
	_ = json.Unmarshal([]byte(`{"dataSources":[{"name":"getHealth","supportedScope":"single"},{"name":"getData","supportedScope":"list"}]}`), &plugin)

	supportedScope, err := findSupportedScope(plugin, dataStreamPlugin{PluginID: "plugin-abc", DataSourceName: "getData"})
	if err != nil || supportedScope != "list" {
		t.Errorf("supported scope = %q, %v, want list", supportedScope, err)
	}

	if _, err := findSupportedScope(plugin, dataStreamPlugin{PluginID: "plugin-abc", DataSourceName: "missing"}); err == nil {
		t.Error("a function the plugin doesn't declare should be an error, not a guess")
	}
}

func TestSupportedScopeDiagnostics(t *testing.T) {
	objectIDs, _ := types.SetValueFrom(context.Background(), types.StringType, []string{"node-a"})
	noObjects := squaredupCache{ObjectIDs: types.SetNull(types.StringType), Scope: jsontypes.NewNormalizedNull()}
	withObjects := squaredupCache{ObjectIDs: objectIDs, Scope: jsontypes.NewNormalizedNull()}

	cases := []struct {
		name           string
		plan           squaredupCache
		supportedScope string
		wantErrors     int
		wantWarnings   int
	}{
		{"objects on a stream that doesn't work on objects", withObjects, "none", 1, 0},
		{"no objects on a stream that doesn't work on objects", noObjects, "none", 0, 0},
		{"objects on a per-object stream", withObjects, "list", 0, 0},
		{"no objects on a per-object stream", noObjects, "single", 0, 1},
	}

	for _, c := range cases {
		c.plan.SupportedScope = types.StringValue(c.supportedScope)
		diags := supportedScopeDiagnostics(c.plan)
		if diags.ErrorsCount() != c.wantErrors || diags.WarningsCount() != c.wantWarnings {
			t.Errorf("%s: got %d errors and %d warnings, want %d and %d", c.name, diags.ErrorsCount(), diags.WarningsCount(), c.wantErrors, c.wantWarnings)
		}
	}

	unknown := withObjects
	unknown.SupportedScope = types.StringUnknown()
	if diags := supportedScopeDiagnostics(unknown); diags.HasError() || diags.WarningsCount() > 0 {
		t.Errorf("nothing can be checked before the data stream is known, got %v", diags)
	}
}
