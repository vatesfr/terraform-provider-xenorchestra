// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	datasources "github.com/vatesfr/terraform-provider-xenorchestra/v2/internal/data_sources"
)

// Ensure XenOrchestraProvider satisfies various provider interfaces.
var _ provider.Provider = &XenOrchestraProvider{}
var _ provider.ProviderWithFunctions = &XenOrchestraProvider{}
var _ provider.ProviderWithEphemeralResources = &XenOrchestraProvider{}
var _ provider.ProviderWithActions = &XenOrchestraProvider{}

// XenOrchestraProvider is the provider implementation.
type XenOrchestraProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// XenOrchestraProviderModel describes the provider data model.
type XenOrchestraProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *XenOrchestraProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	tflog.Info(ctx, "XenOrchestra Terraform Provider - v2")
	resp.TypeName = "xenorchestra"
	resp.Version = p.version
}

func (p *XenOrchestraProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configuration for the Xen Orchestra Terraform provider.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Hostname of the Xen Orchestra server. Can be set via the `XOA_URL` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for Xen Orchestra API. Can be set via the `XOA_USER` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for Xen Orchestra API. Can be set via the `XOA_PASSWORD` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token for Xen Orchestra. Can be set via the `XOA_TOKEN` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Whether to skip SSL certificate verification. Can be set via the `XOA_INSECURE` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *XenOrchestraProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data XenOrchestraProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(applyEnvDefaults(&data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	client, diags := createProviderClient(&data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *XenOrchestraProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *XenOrchestraProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *XenOrchestraProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewPoolDataSource,
		datasources.NewPoolsDataSource,
	}
}

func (p *XenOrchestraProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *XenOrchestraProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &XenOrchestraProvider{
			version: version,
		}
	}
}
