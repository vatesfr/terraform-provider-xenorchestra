package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceXoaResourceSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResourceSetRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceResourceSetRead(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)

	if err != nil {
		return err
	}

	name := d.Get("name").(string)

	rs, err := c.GetResourceSet(client.ResourceSet{Name: name})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	d.SetId(rs.Id)
	d.Set("name", rs.Name)
	return nil
}
