// Copyright (c) Vates
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProviderClient(t *testing.T) {
	t.Run("MissingURL", func(t *testing.T) {
		config := &xenorchestraProviderModel{
			URL:      types.StringNull(),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringValue("test-token"),
			Insecure: types.BoolNull(),
		}
		client, diags := createProviderClient(config)
		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})

	t.Run("MissingCredentials", func(t *testing.T) {
		config := &xenorchestraProviderModel{
			URL:      types.StringValue("http://localhost:8080"),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		client, diags := createProviderClient(config)
		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})

	t.Run("PartialCredentials_UsernameOnly", func(t *testing.T) {
		config := &xenorchestraProviderModel{
			URL:      types.StringValue("http://localhost:8080"),
			Username: types.StringValue("user"),
			Password: types.StringNull(),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		client, diags := createProviderClient(config)
		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})

	t.Run("PartialCredentials_PasswordOnly", func(t *testing.T) {
		config := &xenorchestraProviderModel{
			URL:      types.StringValue("http://localhost:8080"),
			Username: types.StringNull(),
			Password: types.StringValue("pass"),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		client, diags := createProviderClient(config)
		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})

	t.Run("InvalidURL", func(t *testing.T) {
		config := &xenorchestraProviderModel{
			URL:      types.StringValue("://invalid-url"),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringValue("test-token"),
			Insecure: types.BoolNull(),
		}
		client, diags := createProviderClient(config)
		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})

	t.Run("ValidWithToken", func(t *testing.T) {
		config := &xenorchestraProviderModel{
			URL:      types.StringValue("http://localhost:8080"),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringValue("test-token"),
			Insecure: types.BoolNull(),
		}
		client, diags := createProviderClient(config)
		assert.NotNil(t, client)
		assert.False(t, diags.HasError())
	})
}

func TestApplyEnvDefaults(t *testing.T) {
	t.Run("URL", func(t *testing.T) {
		t.Setenv("XOA_URL", "https://env-host:8080")
		config := &xenorchestraProviderModel{
			URL:      types.StringNull(),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		diags := applyEnvDefaults(config)
		require.False(t, diags.HasError())
		assert.Equal(t, "https://env-host:8080", config.URL.ValueString())
	})

	t.Run("DoesNotOverrideExisting", func(t *testing.T) {
		t.Setenv("XOA_URL", "https://env-host:8080")
		config := &xenorchestraProviderModel{
			URL:      types.StringValue("https://manual-host:8080"),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		diags := applyEnvDefaults(config)
		require.False(t, diags.HasError())
		assert.Equal(t, "https://manual-host:8080", config.URL.ValueString())
	})

	t.Run("InvalidInsecure", func(t *testing.T) {
		t.Setenv("XOA_INSECURE", "not-a-bool")
		config := &xenorchestraProviderModel{
			URL:      types.StringNull(),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		diags := applyEnvDefaults(config)
		assert.True(t, diags.HasError())
	})

	t.Run("AllFields", func(t *testing.T) {
		t.Setenv("XOA_URL", "https://env-url")
		t.Setenv("XOA_USER", "env-user")
		t.Setenv("XOA_PASSWORD", "env-pass")
		t.Setenv("XOA_TOKEN", "env-token")
		t.Setenv("XOA_INSECURE", "true")
		config := &xenorchestraProviderModel{
			URL:      types.StringNull(),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		diags := applyEnvDefaults(config)
		require.False(t, diags.HasError())
		assert.Equal(t, "https://env-url", config.URL.ValueString())
		assert.Equal(t, "env-user", config.Username.ValueString())
		assert.Equal(t, "env-pass", config.Password.ValueString())
		assert.Equal(t, "env-token", config.Token.ValueString())
		assert.True(t, config.Insecure.ValueBool())
	})

	t.Run("OnlySomeFields", func(t *testing.T) {
		t.Setenv("XOA_URL", "https://env-url")
		t.Setenv("XOA_TOKEN", "env-token")
		config := &xenorchestraProviderModel{
			URL:      types.StringNull(),
			Username: types.StringNull(),
			Password: types.StringNull(),
			Token:    types.StringNull(),
			Insecure: types.BoolNull(),
		}
		diags := applyEnvDefaults(config)
		require.False(t, diags.HasError())
		assert.Equal(t, "https://env-url", config.URL.ValueString())
		assert.Equal(t, "env-token", config.Token.ValueString())
		assert.Empty(t, config.Username.ValueString())
		assert.Empty(t, config.Password.ValueString())
	})
}
