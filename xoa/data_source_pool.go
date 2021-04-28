package xoa

import (
	"errors"
	"fmt"
	"log"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceXoaPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePoolRead,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"master": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpus": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourcePoolRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)

	pools, err := c.GetPoolByName(nameLabel)

	if err != nil {
		return err
	}

	l := len(pools)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` pools with name `%s`. Pools must be uniquely named to use this data source", l, nameLabel))
	}

	pool := pools[0]

	log.Printf("[DEBUG] Found pool with %+v", pool)
	d.SetId(pool.Id)
	cpus := map[string]string{
		"sockets": fmt.Sprintf("%d", pool.Cpus.Sockets),
		"cores":   fmt.Sprintf("%d", pool.Cpus.Cores),
	}
	if err := d.Set("description", pool.Description); err != nil {
		return err
	}
	err = d.Set("cpus", cpus)
	if err != nil {
		return err
	}

	if err := d.Set("master", pool.Master); err != nil {
		return err
	}
	return nil
}
