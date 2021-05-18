package xoa

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceTags() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}
