package xoa

import "github.com/hashicorp/terraform/helper/schema"

func dataSourceXoaTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTemplateRead,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	return nil
}
