package xoa

import (
	"log"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Elem:     resourceHost(),
			},
			"pool_id": &schema.Schema{
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
	poolId := d.Get("pool_id").(string)
	tags := d.Get("tags").([]interface{})

	pool, err := c.GetPools(client.Pool{Id: poolId})
	if err != nil {
		return err
	}
	// TODO: Add check for the length of pool here
	hosts, err := c.GetSortedHosts(client.Host{Pool: pool[0].Id, Tags: tags}, d.Get("sort_by").(string), d.Get("sort_order").(string))

	log.Printf("[DEBUG] found the following hosts: %s", hosts)
	if err != nil {
		return err
	}

	if err = d.Set("hosts", hostsToMapList(hosts)); err != nil {
		return err
	}
	if err = d.Set("master", pool[0].Master); err != nil {
		return err
	}

	// TODO: This can't be the pool master as this could collide with other data sources
	d.SetId(pool[0].Master)
	return nil
}

func hostsToMapList(hosts []client.Host) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(hosts))
	for _, host := range hosts {
		hostMap := map[string]interface{}{
			"id":         host.Id,
			"name_label": host.NameLabel,
			"pool_id":    host.Pool,
			"tags":       host.Tags,
		}
		result = append(result, hostMap)
	}

	return result
}
