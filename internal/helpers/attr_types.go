// Copyright (c) Vates
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// This file contains helper functions to convert schema attributes to attribute types, which are used in various places across the provider.

// AttrTypesFromSchemaAttributes converts a map of schema attributes to a map of attribute types.
func AttrTypesFromSchemaAttributes(schemaAttrs map[string]schema.Attribute) map[string]attr.Type {
	attrTypes := make(map[string]attr.Type, len(schemaAttrs))

	for name, a := range schemaAttrs {
		attrTypes[name] = a.GetType()
	}

	return attrTypes
}
