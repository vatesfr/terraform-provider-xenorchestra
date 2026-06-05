package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

// ProviderConfig is the configuration for the XenOrchestra provider.
type ProviderConfig struct {
	URL         types.String `tfsdk:"url"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Token       types.String `tfsdk:"token"`
	Insecure    types.Bool   `tfsdk:"insecure"`
	RetryMode   types.String `tfsdk:"retry_mode"`
	RetryMaxTime types.String `tfsdk:"retry_max_time"`
}

// ProviderData contains the configured XO client.
type ProviderData struct {
	Client client.XOClient
	Config ProviderConfig
}

// ProviderSchema returns the schema for the provider configuration.
func ProviderSchema() schema.Schema {
	return schema.Schema{
		Description: "Configuration for the Xen Orchestra Terraform provider.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Hostname of the Xen Orchestra server. Can be set via the XOA_URL environment variable.",
				Required:    true,
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
			"retry_mode": schema.StringAttribute{
				Description: "Retry mode for API requests. Can be 'backoff' or 'none'. Can be set via the XOA_RETRY_MODE environment variable.",
				Optional:    true,
			},
			"retry_max_time": schema.StringAttribute{
				Description: "Maximum time for retry attempts (e.g., '30s', '5m'). Can be set via the XOA_RETRY_MAX_TIME environment variable.",
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

	// Parse retry max time
	var duration time.Duration
	if !config.RetryMaxTime.IsNull() && config.RetryMaxTime.ValueString() != "" {
		var err error
		duration, err = time.ParseDuration(config.RetryMaxTime.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid retry_max_time",
				fmt.Sprintf("Failed to parse retry_max_time: %s. Must be a number followed by ms, s, m, or h (e.g., '30s', '5m')", err.Error()),
			)
			return
		}
	}

	// Map retry mode
	var retryMode client.RetryMode = client.Backoff // default
	if !config.RetryMode.IsNull() {
		switch config.RetryMode.ValueString() {
		case "backoff":
			retryMode = client.Backoff
		case "none":
			retryMode = client.None
		default:
			resp.Diagnostics.AddError(
				"Invalid retry_mode",
				fmt.Sprintf("Unknown retry mode: %s. Must be 'backoff' or 'none'", config.RetryMode.ValueString()),
			)
			return
		}
	}

	// Create XO SDK client configuration
	// Using empty strings for optional fields when not set
	xoConfig := client.Config{
		Url:                config.URL.ValueString(),
		Username:           config.Username.ValueString(),
		Password:           config.Password.ValueString(),
		Token:              config.Token.ValueString(),
		InsecureSkipVerify: config.Insecure.ValueBool(),
		RetryMode:          retryMode,
		RetryMaxTime:       duration,
	}

	// Create client
	c, err := client.NewClient(xoConfig)
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
