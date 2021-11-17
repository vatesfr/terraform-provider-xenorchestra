package xoa

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xo-sdk-go/client"
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
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)

	templateReq := client.Template{
		NameLabel: nameLabel,
		PoolId:    poolId,
	}
	templates, err := c.GetTemplate(templateReq)

	if err != nil {
		return err
	}

	l := len(templates)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` templates with query %+v. Templates must be uniquely named to use this data source", l, templateReq))
	}

	tmpl := templates[0]

	d.SetId(tmpl.Id)
	d.Set("uuid", tmpl.Uuid)
	d.Set("name_label", tmpl.NameLabel)
	return nil
}
