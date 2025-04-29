package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DashboardImageResource{}
	_ resource.ResourceWithConfigure   = &DashboardImageResource{}
	_ resource.ResourceWithImportState = &DashboardImageResource{}
)

func SquaredUpDashboardImageResource() resource.Resource {
	return &DashboardImageResource{}
}

type DashboardImageResource struct {
	client *SquaredUpClient
}

type SquaredUpDashboardImage struct {
	ImageId            types.String `tfsdk:"id"`
	TileId             types.String `tfsdk:"tile_id"`
	DashboardId        types.String `tfsdk:"dashboard_id"`
	WorkspacId         types.String `tfsdk:"workspace_id"`
	ImageBase64DataUri types.String `tfsdk:"image_base64_data_uri"`
	ImageFileName      types.String `tfsdk:"image_file_name"`
}

func (r *DashboardImageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard_image"
}

func (r *DashboardImageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SquaredUp Dashboard Image resource allows you to upload an image to a dashboard tile.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the dashboard image which is the same as the tile ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tile_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the tile where the image will be used.",
				Required:            true,
			},
			"dashboard_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the dashboard which contains the tile where the image will be used.",
				Required:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the workspace which contains the dashboard.",
				Required:            true,
			},
			"image_base64_data_uri": schema.StringAttribute{
				MarkdownDescription: "The base64 data URI of the image.",
				Required:            true,
			},
			"image_file_name": schema.StringAttribute{
				MarkdownDescription: "The file name of the image.",
				Required:            true,
			},
		},
	}
}

func (r *DashboardImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*SquaredUpClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Dashboard Image Resource Configure Type",
			fmt.Sprintf("Expected *SquaredUpClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DashboardImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SquaredUpDashboardImage
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := DashboardImage{
		DataURL: plan.ImageBase64DataUri.ValueString(),
		Metadata: DashboardImageMetadata{
			FileName: plan.ImageFileName.ValueString(),
		},
	}

	err := r.client.UploadDashboardImage(plan.WorkspacId.ValueString(), plan.DashboardId.ValueString(), plan.TileId.ValueString(), &requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Uploading Dashboard Image",
			fmt.Sprintf("Unable to upload dashboard image: %v", err),
		)
		return
	}

	dashboardImage, err := r.client.GetDashboardImage(plan.WorkspacId.ValueString(), plan.DashboardId.ValueString(), plan.TileId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dashboard Image During Create",
			fmt.Sprintf("Unable to read dashboard image: %v", err),
		)
		return
	}

	state := SquaredUpDashboardImage{
		ImageId:            types.StringValue(plan.TileId.ValueString()),
		TileId:             types.StringValue(plan.TileId.ValueString()),
		DashboardId:        types.StringValue(plan.DashboardId.ValueString()),
		WorkspacId:         types.StringValue(plan.WorkspacId.ValueString()),
		ImageBase64DataUri: types.StringValue(dashboardImage.DataURL),
		ImageFileName:      types.StringValue(dashboardImage.Metadata.FileName),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SquaredUpDashboardImage
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardImage, err := r.client.GetDashboardImage(state.WorkspacId.ValueString(), state.DashboardId.ValueString(), state.TileId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dashboard Image",
			fmt.Sprintf("Unable to read dashboard image: %v", err),
		)
		return
	}

	state.ImageBase64DataUri = types.StringValue(dashboardImage.DataURL)
	state.ImageFileName = types.StringValue(dashboardImage.Metadata.FileName)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SquaredUpDashboardImage
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := DashboardImage{
		DataURL: plan.ImageBase64DataUri.ValueString(),
		Metadata: DashboardImageMetadata{
			FileName: plan.ImageFileName.ValueString(),
		},
	}

	err := r.client.UploadDashboardImage(plan.WorkspacId.ValueString(), plan.DashboardId.ValueString(), plan.TileId.ValueString(), &requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Dashboard Image",
			fmt.Sprintf("Unable to update dashboard image: %v", err),
		)
		return
	}

	dashboardImage, err := r.client.GetDashboardImage(plan.WorkspacId.ValueString(), plan.DashboardId.ValueString(), plan.TileId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dashboard Image During Update",
			fmt.Sprintf("Unable to read dashboard image: %v", err),
		)
		return
	}

	plan = SquaredUpDashboardImage{
		ImageId:            types.StringValue(plan.TileId.ValueString()),
		TileId:             types.StringValue(plan.TileId.ValueString()),
		DashboardId:        types.StringValue(plan.DashboardId.ValueString()),
		WorkspacId:         types.StringValue(plan.WorkspacId.ValueString()),
		ImageBase64DataUri: types.StringValue(dashboardImage.DataURL),
		ImageFileName:      types.StringValue(dashboardImage.Metadata.FileName),
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DashboardImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SquaredUpDashboardImage
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDashboardImage(state.WorkspacId.ValueString(), state.DashboardId.ValueString(), state.TileId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dashboard Image",
			fmt.Sprintf("Unable to delete dashboard image: %v", err),
		)
		return
	}
}

func (r *DashboardImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: workspace_id,dashboard_id,tile_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), types.StringValue(idParts[0]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dashboard_id"), types.StringValue(idParts[1]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tile_id"), types.StringValue(idParts[2]))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(idParts[2]))...)

}
