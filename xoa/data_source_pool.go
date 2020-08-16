package xoa

import (
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
			"cpus": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourcePoolRead(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)

	if err != nil {
		return err
	}

	nameLabel := d.Get("name_label").(string)

	pool, err := c.GetPoolByName(nameLabel)

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found pool with %+v", pool)
	d.SetId(pool.Id)
	cpus := map[string]string{
		"sockets": fmt.Sprintf("%d", pool.Cpus.Sockets),
		"cores":   fmt.Sprintf("%d", pool.Cpus.Cores),
	}
	d.Set("description", pool.Description)
	err = d.Set("cpus", cpus)
	if err != nil {
		return err
	}
	return nil
}
