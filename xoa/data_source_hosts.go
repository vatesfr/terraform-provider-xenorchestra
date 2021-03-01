package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

func dataSourceXoaHosts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHostsRead,
		Schema: map[string]*schema.Schema{
			"master": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosts": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
			"pool": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": resourceTags(),
			"sort_by": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"sort_order": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceHostsRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)
	poolLabel := d.Get("pool").(string)
	tags := d.Get("tags").([]interface{})

	pool, err := c.GetPoolByName(poolLabel)
	if err != nil {
		return err
	}
	hosts, err := c.GetHostsByPoolName(client.Host{Pool: pool[0].Id, Tags: tags})
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] found the following hosts: %s", hosts)

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}
	err = d.Set("hosts", hosts)
	if err != nil {
		log.Printf("[DEBUG] failed setting hosts: %s", err.Error())
		return err
	}
	err = d.Set("master", pool[0].Master)
	d.SetId(pool[0].Master)

	if err != nil {
		log.Printf("[DEBUG] failed setting master id: %s", err.Error())
		return err
	}
	return nil
}
