package xoa

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceHost() *schema.Resource {
	return &schema.Resource{
		Schema: resourceHostSchema(),
	}
}

func resourceHostSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"name_label": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"pool": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"tags": resourceTags(),
	}
}
