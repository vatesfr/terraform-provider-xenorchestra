package xoa

import (
	"errors"
	"fmt"

	"github.com/vatesfr/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXoaResourceSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResourceSetRead,
		Description: `Provides information about a resource set.

**NOTE:** If there are multiple resource sets with the same name
Terraform will fail. Ensure that your resource set names are unique when
using the data source.`,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource set to look up.",
			},
		},
	}
}

func dataSourceResourceSetRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	name := d.Get("name").(string)

	resourceSets, err := c.GetResourceSet(client.ResourceSet{Name: name})

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
