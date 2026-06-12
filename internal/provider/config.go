// Copyright (c) Vates
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v2config "github.com/vatesfr/xenorchestra-go-sdk/pkg/config"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

// createProviderClient configures the provider client.
func createProviderClient(config *XenOrchestraProviderModel) (*v2.XOClient, diag.Diagnostics) {
	var diags diag.Diagnostics

	if config.URL.IsNull() || config.URL.IsUnknown() || config.URL.ValueString() == "" {
		diags.AddError(
			"Missing Xen Orchestra URL",
			"Set provider argument 'url' or define the XOA_URL environment variable.",
		)
		return nil, diags
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
	c, err := v2.New(xoConfig)
	if err != nil {
		diags.AddError(
			"Failed to create Xen Orchestra client",
			fmt.Sprintf("Error: %s", err.Error()),
		)
		return nil, diags
	}

	xoClient := c.(*v2.XOClient)
	return xoClient, diags
}

func applyEnvDefaults(config *XenOrchestraProviderModel) diag.Diagnostics {
	var diags diag.Diagnostics
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
				diags.AddError(
					"Invalid XOA_INSECURE value",
					fmt.Sprintf("XOA_INSECURE must be a valid boolean value, got %q.", v),
				)
				return diags
			}
			config.Insecure = types.BoolValue(insecure)
		}
	}
	return diags
}
