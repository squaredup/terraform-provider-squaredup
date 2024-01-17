package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &squaredupProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &squaredupProvider{
			version: version,
		}
	}
}

type squaredupProvider struct {
	version string
}

type squaredupProviderModel struct {
	Region types.String `tfsdk:"region"`
	APIKey types.String `tfsdk:"api_key"`
}

func (p *squaredupProvider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "squaredup"
	resp.Version = p.version
}

func (p *squaredupProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Squaredup",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Region of your SquaredUp instance. May also be set via the SQUAREDUP_REGION environment variable.",
				Optional:    true,
				Validators: []validator.String{stringvalidator.OneOf(
					"us",
					"eu",
				)},
			},
			"api_key": schema.StringAttribute{
				Description: "API Key for SquaredUp API. May also be set via the SQUAREDUP_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *squaredupProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	var config squaredupProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Region.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Unknown SquaredUp Region",
			"The provider cannot create the SquaredUp API client as there is an unknown configuration value for the SquaredUp API Region. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SQUAREDUP_REGION environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown SquaredUp API Key",
			"The provider cannot create the SquaredUp API client as there is an unknown configuration value for the SquaredUp API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the SQUAREDUP_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	region := os.Getenv("SQUAREDUP_REGION")
	apiKey := os.Getenv("SQUAREDUP_API_KEY")

	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if region == "" {
		region = "us"
		resp.Diagnostics.AddAttributeWarning(
			path.Root("region"),
			"Missing SquaredUp Region",
			"Region not set in configuration or environment variable. Defaulting to US. "+
				"Set the region value in the configuration or use the SQUAREDUP_REGION environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing SquaredUp API Key",
			"The provider cannot create the SquaredUp API client as there is a missing or empty value for the SquaredUp API Key. "+
				"Set the API Key value in the configuration or use the SQUAREDUP_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "squaredup_api_key", apiKey)
	tflog.MaskFieldValuesWithFieldKeys(ctx, "squaredup_api_key")

	client, err := NewSquaredUpClient(region, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create SquaredUp API Client",
			"An unexpected error occurred while creating the SquaredUp API client. "+
				"Please check the configuration and try again. If the error persists, please open an issue on GitHub. "+
				err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *squaredupProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		SquaredupDataSourcesDataSource,
		SquaredUpDataStreams,
		SquaredUpNodes,
	}
}

func (p *squaredupProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		SquaredupDataSourceResource,
		SquaredupWorkspaceResource,
		SquaredUpDashboardResource,
		SquaredUpOpenAccessResource,
	}
}
