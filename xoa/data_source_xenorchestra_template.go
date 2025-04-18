package xoa

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

func dataSourceXoaTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTemplateRead,
		Description: `Provides information about a VM template that can be used for creating new VMs.

**Note:** If there are multiple templates that match terraform will fail.
Ensure that your name_label and pool_id identify a unique template.`,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the template to look up.",
				Required:    true,
			},
			"uuid": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The uuid of the template.",
				Computed:    true,
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The id of the pool that the template belongs to.",
				Optional:    true,
			},
		},
	}
}

func dataSourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)

	templateReq := client.Template{
		NameLabel: nameLabel,
		PoolId:    poolId,
	}
	templates, err := c.V1Client().GetTemplate(templateReq)

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
