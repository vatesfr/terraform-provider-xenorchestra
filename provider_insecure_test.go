package main

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vatesfr/terraform-provider-xenorchestra/xoa"
)

// configureInsecure runs the provider configuration the same way terraform does
// (defaults and DefaultFuncs applied) and returns the resulting `insecure` value
// without building a client.
func configureInsecure(t *testing.T, config map[string]interface{}) bool {
	t.Helper()

	provider := xoa.Provider()

	var insecure bool
	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		insecure = d.Get("insecure").(bool)
		return nil, nil
	}

	diags := provider.Configure(context.Background(), terraform.NewResourceConfigRaw(config))
	require.False(t, diags.HasError(), "failed to configure provider: %v", diags)

	return insecure
}

func TestProviderInsecureFromEnvironment(t *testing.T) {
	for _, tt := range []struct {
		name     string
		env      string
		config   map[string]interface{}
		expected bool
	}{
		{
			name:     "unset environment variable defaults to secure",
			env:      "",
			expected: false,
		},
		{
			name:     "XOA_INSECURE=true",
			env:      "true",
			expected: true,
		},
		{
			name:     "XOA_INSECURE=false",
			env:      "false",
			expected: false,
		},
		{
			name:     "attribute takes precedence over the environment variable",
			env:      "true",
			config:   map[string]interface{}{"insecure": false},
			expected: false,
		},
		{
			name:     "attribute is honored without the environment variable",
			env:      "",
			config:   map[string]interface{}{"insecure": true},
			expected: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XOA_INSECURE", tt.env)

			assert.Equal(t, tt.expected, configureInsecure(t, tt.config), "insecure")
		})
	}
}
