package xoa

import (
	"errors"
	"fmt"

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
	c := m.(*client.Client)

	name := d.Get("name").(string)

	resourceSets, err := c.GetResourceSet(client.ResourceSet{Name: name})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}
	l := len(resourceSets)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` resource sets with name `%s`. Resource sets must be uniquely named to use this data source. Rename the conflicting resource set and try again.", l, name))
	}

	rs := resourceSets[0]
	d.SetId(rs.Id)
	d.Set("name", rs.Name)
	return nil
}
