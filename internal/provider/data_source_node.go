package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &squaredupNodes{}
	_ datasource.DataSourceWithConfigure = &squaredupNodes{}
)

func SquaredUpNodes() datasource.DataSource {
	return &squaredupNodes{}
}

type squaredupNodes struct {
	client *SquaredUpClient
}

type squaredupNodesResponse struct {
	NodeProperties []squaredupNodesProperties `tfsdk:"node_properties"`
	DataSourceID   types.String               `tfsdk:"data_source_id"`
	NodeName       types.String               `tfsdk:"node_name"`
	NodeSourceID   types.String               `tfsdk:"node_source_id"`
	AllowNoData    types.Bool                 `tfsdk:"allow_no_data"`
}

type squaredupNodesProperties struct {
	ID          types.String `tfsdk:"id"`
	SourceName  types.String `tfsdk:"source_name"`
	DisplayName types.String `tfsdk:"display_name"`
	SourceID    types.String `tfsdk:"source_id"`
	Type        types.String `tfsdk:"type"`
}

func (d *squaredupNodes) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nodes"
}

func (d *squaredupNodes) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"node_properties": schema.ListNestedAttribute{
				MarkdownDescription: "Node Properties",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"source_name":  schema.StringAttribute{Computed: true},
						"display_name": schema.StringAttribute{Computed: true},
						"source_id":    schema.StringAttribute{Computed: true},
						"type":         schema.StringAttribute{Computed: true},
					},
				},
			},
			"data_source_id": schema.StringAttribute{
				MarkdownDescription: "Data Source ID",
				Required:            true,
			},
			"node_name": schema.StringAttribute{
				MarkdownDescription: "Node Name",
				Optional:            true,
			},
			"node_source_id": schema.StringAttribute{
				MarkdownDescription: "Node Source ID",
				Optional:            true,
			},
			"allow_no_data": schema.BoolAttribute{
				MarkdownDescription: "If true, the data source will return an empty list if its unable to find the node.",
				Optional:            true,
			},
		},
	}
}

func (d *squaredupNodes) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *squaredupNodes) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state squaredupNodesResponse
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.NodeName.ValueString() != "" && state.NodeSourceID.ValueString() != "" {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Both node_name and node_source_id cannot be used at the same time",
		)
		return
	}

	nodes, err := d.client.GetNodes(state.DataSourceID.ValueString(), state.NodeName.ValueString(), state.NodeSourceID.ValueString(), state.AllowNoData.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Retrieve Nodes",
			err.Error(),
		)
		return
	}

	var NodeProperties []squaredupNodesProperties
	for _, node := range nodes {
		nodeProperties := squaredupNodesProperties{
			ID:          types.StringValue(node.ID),
			SourceName:  types.StringValue(node.SourceName[0]),
			DisplayName: types.StringValue(node.DisplayName[0]),
			SourceID:    types.StringValue(node.SourceID[0]),
			Type:        types.StringValue(node.Type[0]),
		}
		NodeProperties = append(NodeProperties, nodeProperties)
	}
	state.NodeProperties = NodeProperties

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
