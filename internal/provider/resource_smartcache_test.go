package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/pborman/uuid"
)

func smartCacheTestConfig(displayName string, frequencyMinutes string, enabled string) string {
	return providerConfig + `
data "squaredup_datasources" "sample_data" {
	data_source_name = "Sample Data"
}

resource "squaredup_datasource" "sample_data_source" {
	display_name     = "Sample Data - SmartCache Test"
	data_source_name = data.squaredup_datasources.sample_data.plugins[0].display_name
}

resource "squaredup_workspace" "smartcache_workspace" {
	display_name      = "SmartCache Test - ` + displayName + `"
	datasources_links = [squaredup_datasource.sample_data_source.id]
}

data "squaredup_data_streams" "database_free_space" {
	data_source_id              = data.squaredup_datasources.sample_data.plugins[0].id
	data_stream_definition_name = "database-free-space"
}

resource "squaredup_smartcache" "test" {
	display_name      = "` + displayName + `"
	data_source_id    = squaredup_datasource.sample_data_source.id
	data_stream_id    = data.squaredup_data_streams.database_free_space.data_streams[0].id
	supported_scope   = "list"
	scope = jsonencode({
		workspaceId = squaredup_workspace.smartcache_workspace.id
		query       = "g.V().has('type', 'sample-database')"
	})
	frequency_minutes = ` + frequencyMinutes + `
	enabled           = ` + enabled + `
}
`
}

func TestAccResourceSmartCache(t *testing.T) {
	// Cache names are capped at 24 characters
	suffix := uuid.NewRandom().String()[:8]
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: smartCacheTestConfig("TF Cache "+suffix, "60", "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "display_name", "TF Cache "+suffix),
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "frequency_minutes", "60"),
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "sync_mode", "full"),
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "retain_mirrors", "1"),
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "timeframe", "last24hours"),
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("squaredup_smartcache.test", "id"),
					resource.TestCheckResourceAttrSet("squaredup_smartcache.test", "table_name"),
				),
			},
			// Import Test
			{
				ResourceName:            "squaredup_smartcache.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"scope"},
			},
			// Update Test
			{
				Config: smartCacheTestConfig("TF Cache Renamed "+suffix, "120", "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "display_name", "TF Cache Renamed "+suffix),
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "frequency_minutes", "120"),
					resource.TestCheckResourceAttr("squaredup_smartcache.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestSmartCacheStateMapsIncrementalStorage(t *testing.T) {
	columns := []string{"id"}
	lastModified := "updatedAt"
	retention := int64(90)
	read := &SmartCacheRead{
		ID:               "smartcachedefinition-abc",
		Name:             "Costs",
		FrequencyMinutes: 60,
		Enabled:          true,
		Steps: []SmartCacheStep{{
			DataStream: SmartCacheDataStream{
				ID:               "datastream-abc",
				PluginConfigID:   "config-abc",
				DataSourceConfig: map[string]interface{}{"region": "eu"},
			},
			Timeframe:      "last7days",
			SupportedScope: "none",
		}},
	}
	read.Steps[0].SmartCacheMapping.TableName = "costs_table"
	read.Steps[0].SmartCacheMapping.SyncMode = "incremental"
	read.Steps[0].SmartCacheMapping.UniqueIdentifierColumns = &columns
	read.Steps[0].SmartCacheMapping.LastModifiedColumn = &lastModified
	read.Steps[0].SmartCacheMapping.RetentionPeriodDays = &retention

	prior := squaredupSmartCache{Scope: jsontypes.NewNormalizedValue(`{"query":"g.V()"}`)}

	state, diags := smartCacheState(context.Background(), read, prior)
	if diags.HasError() {
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
	if !state.RetainMirrors.IsNull() || state.RetentionPeriodDays.ValueInt64() != 90 {
		t.Errorf("storage not mapped: retain_mirrors %v, retention_period_days %v", state.RetainMirrors, state.RetentionPeriodDays)
	}
	if state.LastModifiedColumn.ValueString() != "updatedAt" || len(state.UniqueIdentifierColumns.Elements()) != 1 {
		t.Errorf("incremental columns not mapped: %v, %v", state.LastModifiedColumn, state.UniqueIdentifierColumns)
	}
}

func TestSmartCacheStateTreatsClearedSettingsAsNull(t *testing.T) {
	empty := ""
	columns := []string{}
	read := &SmartCacheRead{ID: "smartcachedefinition-abc", DisplayName: "Costs", Steps: []SmartCacheStep{{}}}
	read.Steps[0].SmartCacheMapping.SyncMode = "incremental"
	read.Steps[0].SmartCacheMapping.LastModifiedColumn = &empty
	read.Steps[0].SmartCacheMapping.UniqueIdentifierColumns = &columns

	state, diags := smartCacheState(context.Background(), read, squaredupSmartCache{})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !state.LastModifiedColumn.IsNull() || !state.UniqueIdentifierColumns.IsNull() {
		t.Errorf("cleared settings should be null, got %v, %v", state.LastModifiedColumn, state.UniqueIdentifierColumns)
	}
	if state.FrequencyMinutes.ValueInt64() != 1440 {
		t.Errorf("a missing frequency should read as daily, got %d", state.FrequencyMinutes.ValueInt64())
	}
}

func TestSmartCacheStorageClearsRemovedSettingsOnUpdate(t *testing.T) {
	plan := squaredupSmartCache{
		SyncMode:                types.StringValue("incremental"),
		RetainMirrors:           types.Int64Unknown(),
		RetentionPeriodDays:     types.Int64Value(30),
		LastModifiedColumn:      types.StringNull(),
		UniqueIdentifierColumns: types.ListNull(types.StringType),
	}

	created, _ := smartCacheStorage(context.Background(), plan, false)
	if created.LastModifiedColumn != nil || created.UniqueIdentifierColumns != nil || created.RetainMirrors != nil {
		t.Errorf("create should leave unset settings to the server defaults, got %+v", created)
	}

	updated, _ := smartCacheStorage(context.Background(), plan, true)
	if updated.LastModifiedColumn == nil || *updated.LastModifiedColumn != "" {
		t.Errorf("update should clear last_modified_column, got %v", updated.LastModifiedColumn)
	}
	if updated.UniqueIdentifierColumns == nil || len(*updated.UniqueIdentifierColumns) != 0 {
		t.Errorf("update should clear unique_identifier_columns, got %v", updated.UniqueIdentifierColumns)
	}
}
