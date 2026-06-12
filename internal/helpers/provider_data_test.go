// Copyright (c) Vates
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
	"github.com/stretchr/testify/assert"
)

func TestProviderDataToXOClient_NilInput(t *testing.T) {
	var diags diag.Diagnostics

	client := ProviderDataToXOClient(nil, &diags)

	assert.Nil(t, client)
	assert.False(t, diags.HasError())
}

func TestProviderDataToXOClient_ValidClient(t *testing.T) {
	var diags diag.Diagnostics

	expectedClient := &v2.XOClient{}
	client := ProviderDataToXOClient(expectedClient, &diags)

	assert.Same(t, expectedClient, client)
	assert.False(t, diags.HasError())
}

func TestProviderDataToXOClient_WrongType(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var diags diag.Diagnostics

		client := ProviderDataToXOClient("not a client", &diags)

		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})

	t.Run("Int", func(t *testing.T) {
		var diags diag.Diagnostics

		client := ProviderDataToXOClient(42, &diags)

		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})

	t.Run("Map", func(t *testing.T) {
		var diags diag.Diagnostics

		client := ProviderDataToXOClient(map[string]any{"key": "value"}, &diags)

		assert.Nil(t, client)
		assert.True(t, diags.HasError())
	})
}
