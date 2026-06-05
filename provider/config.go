package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v2config "github.com/vatesfr/xenorchestra-go-sdk/pkg/config"
	"github.com/vatesfr/xenorchestra-go-sdk/v2/client"
)

// ProviderConfig is the configuration for the XenOrchestra provider.
type ProviderConfig struct {
	URL      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// ProviderData contains the configured XO client.
type ProviderData struct {
	Client *client.Client
	Config ProviderConfig
}

// ProviderSchema returns the schema for the provider configuration.
func ProviderSchema() schema.Schema {
	return schema.Schema{
		Description: "Configuration for the Xen Orchestra Terraform provider.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Hostname of the Xen Orchestra server. If omitted, XOA_URL is used.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for Xen Orchestra API. Can be set via the XOA_USER environment variable.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Xen Orchestra API. Can be set via the XOA_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"token": schema.StringAttribute{
				Description: "API token for Xen Orchestra. Can be set via the XOA_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Whether to skip SSL certificate verification. Can be set via the XOA_INSECURE environment variable.",
				Optional:    true,
			},
		},
	}
}

// ConfigureProvider configures the provider client.
func ConfigureProvider(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Debug(ctx, "Configuring XenOrchestra provider")

	var config ProviderConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	applyEnvDefaults(&config, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.URL.IsNull() || config.URL.IsUnknown() || config.URL.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing Xen Orchestra URL",
			"Set provider argument 'url' or define the XOA_URL environment variable.",
		)
		return
	}

	// Create XO SDK client configuration
	// Using empty strings for optional fields when not set
	xoConfig := &v2config.Config{
		Url:                config.URL.ValueString(),
		Username:           config.Username.ValueString(),
		Password:           config.Password.ValueString(),
		Token:              config.Token.ValueString(),
		InsecureSkipVerify: config.Insecure.ValueBool(),
	}
	// Create client
	c, err := client.New(xoConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create Xen Orchestra client",
			fmt.Sprintf("Error: %s", err.Error()),
		)
		return
	}

	// Set the client in the provider data
	providerData := &ProviderData{
		Client: c,
		Config: config,
	}

	resp.ResourceData = providerData
	resp.DataSourceData = providerData
}

func applyEnvDefaults(config *ProviderConfig, resp *provider.ConfigureResponse) {
	if (config.URL.IsNull() || config.URL.IsUnknown()) && os.Getenv("XOA_URL") != "" {
		config.URL = types.StringValue(os.Getenv("XOA_URL"))
	}

	if (config.Username.IsNull() || config.Username.IsUnknown()) && os.Getenv("XOA_USER") != "" {
		config.Username = types.StringValue(os.Getenv("XOA_USER"))
	}

	if (config.Password.IsNull() || config.Password.IsUnknown()) && os.Getenv("XOA_PASSWORD") != "" {
		config.Password = types.StringValue(os.Getenv("XOA_PASSWORD"))
	}

	if (config.Token.IsNull() || config.Token.IsUnknown()) && os.Getenv("XOA_TOKEN") != "" {
		config.Token = types.StringValue(os.Getenv("XOA_TOKEN"))
	}

	if config.Insecure.IsNull() || config.Insecure.IsUnknown() {
		if v := os.Getenv("XOA_INSECURE"); v != "" {
			insecure, err := strconv.ParseBool(v)
			if err != nil {
				resp.Diagnostics.AddError(
					"Invalid XOA_INSECURE value",
					fmt.Sprintf("XOA_INSECURE must be a valid boolean value, got %q.", v),
				)
				return
			}
			config.Insecure = types.BoolValue(insecure)
		}
	}
}
