package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceXoaTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTemplateRead,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"uuid": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	c, err := client.NewClient()

	if err != nil {
		return err
	}

	templateName := d.Get("name_label").(string)

	tmpl, err := c.GetTemplate(templateName)

	if err != nil {
		return err
	}

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	d.SetId(tmpl.Id)
	d.Set("uuid", tmpl.Uuid)
	d.Set("name_label", tmpl.NameLabel)
	return nil
}
