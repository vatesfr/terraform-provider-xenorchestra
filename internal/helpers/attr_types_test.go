// Copyright (c) Vates
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAttrTypesFromSchemaAttributes_EmptyMap(t *testing.T) {
	schemaAttrs := map[string]schema.Attribute{}

	result := AttrTypesFromSchemaAttributes(schemaAttrs)

	assert.Empty(t, result)
}

func TestAttrTypesFromSchemaAttributes_SingleString(t *testing.T) {
	schemaAttrs := map[string]schema.Attribute{
		"name": schema.StringAttribute{},
	}

	result := AttrTypesFromSchemaAttributes(schemaAttrs)

	expected := map[string]attr.Type{
		"name": types.StringType,
	}

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestAttrTypesFromSchemaAttributes_MultipleTypes(t *testing.T) {
	schemaAttrs := map[string]schema.Attribute{
		"name":     schema.StringAttribute{},
		"enabled":  schema.BoolAttribute{},
		"count":    schema.Int64Attribute{},
		"amount":   schema.Float64Attribute{},
		"tags":     schema.ListAttribute{ElementType: types.StringType},
		"metadata": schema.MapAttribute{ElementType: types.StringType},
	}

	result := AttrTypesFromSchemaAttributes(schemaAttrs)

	assert.Len(t, result, len(schemaAttrs))

	if diff := cmp.Diff(types.StringType, result["name"]); diff != "" {
		t.Errorf("name type mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(types.BoolType, result["enabled"]); diff != "" {
		t.Errorf("enabled type mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(types.Int64Type, result["count"]); diff != "" {
		t.Errorf("count type mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(types.Float64Type, result["amount"]); diff != "" {
		t.Errorf("amount type mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(types.ListType{ElemType: types.StringType}, result["tags"]); diff != "" {
		t.Errorf("tags type mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(types.MapType{ElemType: types.StringType}, result["metadata"]); diff != "" {
		t.Errorf("metadata type mismatch (-want +got):\n%s", diff)
	}
}

func TestAttrTypesFromSchemaAttributes_NestedAttribute(t *testing.T) {
	schemaAttrs := map[string]schema.Attribute{
		"config": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"key": schema.StringAttribute{},
			},
		},
	}

	result := AttrTypesFromSchemaAttributes(schemaAttrs)

	expected := map[string]attr.Type{
		"config": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key": types.StringType,
			},
		},
	}

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestAttrTypesFromSchemaAttributes_PreservesAllKeys(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"}
	schemaAttrs := make(map[string]schema.Attribute, len(keys))
	for _, k := range keys {
		schemaAttrs[k] = schema.StringAttribute{}
	}

	result := AttrTypesFromSchemaAttributes(schemaAttrs)

	assert.Len(t, result, len(keys))

	for _, k := range keys {
		assert.Contains(t, result, k)
	}
}
