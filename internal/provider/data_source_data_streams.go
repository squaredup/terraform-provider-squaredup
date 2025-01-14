package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &squaredupDataStream{}
	_ datasource.DataSourceWithConfigure = &squaredupDataStream{}
)

func SquaredUpDataStreams() datasource.DataSource {
	return &squaredupDataStream{}
}

type squaredupDataStream struct {
	client *SquaredUpClient
}

type squaredupDataStreams struct {
	DataStreams    []squaredupDataSourceDataStreams `tfsdk:"data_streams"`
	DataSourceID   types.String                     `tfsdk:"data_source_id"`
	DataStreamName types.String                     `tfsdk:"data_stream_definition_name"`
}

type squaredupDataSourceDataStreams struct {
	DisplayName    types.String `tfsdk:"display_name"`
	DataSourceName types.String `tfsdk:"data_source_name"`
	Id             types.String `tfsdk:"id"`
	DefinitionName types.String `tfsdk:"definition_name"`
}

func (d *squaredupDataStream) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_streams"
}

func (d *squaredupDataStream) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"data_streams": schema.ListNestedAttribute{
				MarkdownDescription: "Data Streams are used to query third party APIs and SquaredUp visualizes the results",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name":     schema.StringAttribute{Computed: true},
						"data_source_name": schema.StringAttribute{Computed: true},
						"id":               schema.StringAttribute{Computed: true},
						"definition_name":  schema.StringAttribute{Computed: true},
					},
				},
			},
			"data_source_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the data source to get data streams for",
			},
			"data_stream_definition_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the data stream definition to get particular data stream details",
			},
		},
	}
}

func (d *squaredupDataStream) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state squaredupDataStreams
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataStreams, err := d.client.GetDataStreams(state.DataSourceID.ValueString(), state.DataStreamName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get data streams",
			err.Error(),
		)
		return
	}

	for _, dataStream := range dataStreams {
		dataStreamState := squaredupDataSourceDataStreams{
			DisplayName:    types.StringValue(dataStream.DisplayName),
			DataSourceName: types.StringValue(dataStream.DataSourceName),
			DefinitionName: types.StringValue(dataStream.Definition.Name),
			Id:             types.StringValue(dataStream.ID),
		}
		state.DataStreams = append(state.DataStreams, dataStreamState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *squaredupDataStream) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to cast provider data to SquaredUpClient",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}
