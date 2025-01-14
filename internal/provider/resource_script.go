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
)

var (
	_ resource.Resource                = &ScriptResource{}
	_ resource.ResourceWithConfigure   = &ScriptResource{}
	_ resource.ResourceWithImportState = &ScriptResource{}
)

func SquaredUpScriptResource() resource.Resource {
	return &ScriptResource{}
}

type ScriptResource struct {
	client *SquaredUpClient
}

type squaredupScript struct {
	ScriptID    types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	ScriptType  types.String `tfsdk:"script_type"`
	Script      types.String `tfsdk:"script"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *ScriptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_script"
}

func (r *ScriptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SquaredUp Script",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the script",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name for the script",
				Required:            true,
			},
			"script_type": schema.StringAttribute{
				MarkdownDescription: "Type of script. Must be one of: tileDataJS, monitorConditionJS",
				Required:            true,
				Validators: []validator.String{stringvalidator.OneOf(
					"tileDataJS",
					"monitorConditionJS",
				)},
			},
			"script": schema.StringAttribute{
				MarkdownDescription: "Contents of the script",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "The last updated date of the script",
				Computed:            true,
			},
		},
	}
}

func (r *ScriptResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := squaredupScript{}
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	scriptPayload := Script{
		DisplayName: plan.DisplayName.ValueString(),
		ScriptType:  plan.ScriptType.ValueString(),
		Config: ScriptConfig{
			Src: plan.Script.ValueString(),
		},
	}

	script, err := r.client.CreateScript(scriptPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create script",
			err.Error(),
		)
		return
	}

	state := squaredupScript{
		ScriptID:    types.StringValue(script.ID),
		DisplayName: types.StringValue(script.DisplayName),
		ScriptType:  types.StringValue(strings.Split(script.SubType, ".")[1]),
		Script:      types.StringValue(script.Config.Src),
		LastUpdated: types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state squaredupScript
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	script, err := r.client.GetScript(state.ScriptID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read script",
			err.Error(),
		)
		return
	}

	state = squaredupScript{
		ScriptID:    types.StringValue(script.ID),
		DisplayName: types.StringValue(script.DisplayName),
		ScriptType:  types.StringValue(strings.Split(script.SubType, ".")[1]),
		Script:      types.StringValue(script.Config.Src),
		LastUpdated: types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan squaredupScript
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		return
	}

	var state squaredupScript
	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		return
	}

	if plan.ScriptType.ValueString() != state.ScriptType.ValueString() {
		resp.Diagnostics.AddError(
			"Script type cannot be updated",
			"Script type is not allowed to be updated from "+state.ScriptType.ValueString()+" to "+plan.ScriptType.ValueString()+". Please delete and recreate the resource.",
		)
		return
	}

	scriptPayload := Script{
		DisplayName: plan.DisplayName.ValueString(),
		ScriptType:  plan.ScriptType.ValueString(),
		Config: ScriptConfig{
			Src: plan.Script.ValueString(),
		},
	}

	err := r.client.UpdateScript(plan.ScriptID.ValueString(), scriptPayload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update script",
			err.Error(),
		)
		return
	}

	script, err := r.client.GetScript(plan.ScriptID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read script",
			err.Error(),
		)
		return
	}

	state = squaredupScript{
		ScriptID:    types.StringValue(script.ID),
		DisplayName: types.StringValue(script.DisplayName),
		ScriptType:  types.StringValue(strings.Split(script.SubType, ".")[1]),
		Script:      types.StringValue(script.Config.Src),
		LastUpdated: types.StringValue(time.Now().Format(time.RFC850)),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state squaredupScript
	diags := req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics = diags
		return
	}

	err := r.client.DeleteScript(state.ScriptID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete script",
			err.Error(),
		)
		return
	}
}

func (r *ScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
