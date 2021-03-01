package xoa

import (
	"errors"
	"fmt"
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceXoaHost() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceHostRead,
		Schema: resourceHostSchema(),
	}
}

func dataSourceHostRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)
	nameLabel := d.Get("name_label").(string)
	hosts, err := c.GetHostByName(nameLabel)

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

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
