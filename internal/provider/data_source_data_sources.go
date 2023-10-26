package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &squaredupLatestDataSource{}
	_ datasource.DataSourceWithConfigure = &squaredupLatestDataSource{}
)

func SquaredupDataSourcesDataSource() datasource.DataSource {
	return &squaredupLatestDataSource{}
}

type squaredupLatestDataSource struct {
	client *SquaredUpClient
}

type squaredupDataSourceModel struct {
	Plugins        []squaredupPluginModel `tfsdk:"plugins"`
	DataSourceName types.String           `tfsdk:"data_source_name"`
}

type squaredupPluginModel struct {
	Category    types.String `tfsdk:"category"`
	Description types.String `tfsdk:"description"`
	Author      types.String `tfsdk:"author"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Version     types.String `tfsdk:"version"`
	OnPrem      types.Bool   `tfsdk:"on_prem"`
	DisplayName types.String `tfsdk:"display_name"`
	PluginID    types.String `tfsdk:"id"`
}

func (d *squaredupLatestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datasources"
}

func (d *squaredupLatestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data Sources are used to query third party APIs and SquaredUp visualizes the results",
		Attributes: map[string]schema.Attribute{
			"data_source_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the data source. If not specified, all data sources will be returned.",
			},
			"plugins": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"category":     schema.StringAttribute{Computed: true},
						"description":  schema.StringAttribute{Computed: true},
						"author":       schema.StringAttribute{Computed: true},
						"last_updated": schema.StringAttribute{Computed: true},
						"version":      schema.StringAttribute{Computed: true},
						"on_prem":      schema.BoolAttribute{Computed: true},
						"display_name": schema.StringAttribute{Computed: true},
						"id":           schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *squaredupLatestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state squaredupDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plugins, err := d.client.GetLatestDataSources(state.DataSourceName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error making API request to fetch latest Data Sources",
			err.Error(),
		)
		return
	}

	for _, plugin := range plugins {
		pluginsState := squaredupPluginModel{
			Category:    types.StringValue(plugin.Category),
			Description: types.StringValue(plugin.Description),
			Author:      types.StringValue(plugin.Author),
			LastUpdated: types.StringValue(plugin.LastUpdated),
			Version:     types.StringValue(plugin.Version),
			OnPrem:      types.BoolValue(plugin.OnPrem),
			DisplayName: types.StringValue(plugin.DisplayName),
			PluginID:    types.StringValue(plugin.PluginID),
		}
		state.Plugins = append(state.Plugins, pluginsState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *squaredupLatestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}
