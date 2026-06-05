// Package provider contains the Terraform provider implementation for Xen Orchestra.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// XenOrchestraProvider is the provider implementation.
type XenOrchestraProvider struct {
	version string
}

// NewProvider creates a new XenOrchestraProvider instance with version.
func NewProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &XenOrchestraProvider{
			version: version,
		}
	}
}

// New creates a new XenOrchestraProvider instance (for internal use).
func New(version string) provider.Provider {
	return &XenOrchestraProvider{
		version: version,
	}
}

// Metadata returns the provider type name and version.
func (p *XenOrchestraProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	tflog.Info(ctx, "XenOrchestra Terraform Provider - v2")
	
	resp.TypeName = "xenorchestra"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration.
func (p *XenOrchestraProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = ProviderSchema()
}

// Configure configures the provider client.
func (p *XenOrchestraProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	ConfigureProvider(ctx, req, resp)
}

// Resources returns the provider's resource implementations.
// Empty for now - will be added in later phases
func (p *XenOrchestraProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

// DataSources returns the provider's data source implementations.
// Empty for now - will be added in later phases
func (p *XenOrchestraProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
