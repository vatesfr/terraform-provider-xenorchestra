package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceXoaHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHostRead,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceHostRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)
	nameLabel := d.Get("name_label").(string)
	hosts, err := c.GetHostByName(nameLabel)

	if err != nil {
		return err
	}
	l := len(hosts)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` hosts with name_label `%s`. Hosts must be uniquely named to use this data source", l, nameLabel))
	}

	d.SetId(hosts[0].Id)

	return nil
}
