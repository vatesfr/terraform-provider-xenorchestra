package xoa

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceTags() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "The tags (labels) applied to the given entity.",
	}
}
