package xoa

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaHosts() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceHostsRead,
		Description: "Use this data source to filter Xenorchestra hosts by certain criteria (name_label, tags) for use in other resources.",
		Schema: map[string]*schema.Schema{
			"master": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The primary host of the pool.",
			},
			"hosts": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        resourceHost(),
				Description: "The resulting hosts after applying the argument filtering.",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The pool id used to filter the resulting hosts by.",
			},
			"tags": resourceTags(),
			"sort_by": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The host field to sort the results by (id and name_label are supported).",
			},
			"sort_order": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Valid options are `asc` or `desc` and sort order is applied to `sort_by` argument.",
			},
		},
	}
}

func dataSourceHostsRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)
	poolId := d.Get("pool_id").(string)
	tags := d.Get("tags").(*schema.Set).List()

	pool, err := c.GetPools(client.Pool{Id: poolId})
	if err != nil {
		return err
	}
	searchHost := client.Host{
		Pool: pool[0].Id,
		Tags: tags}
	hosts, err := c.GetSortedHosts(searchHost, d.Get("sort_by").(string), d.Get("sort_order").(string))

	log.Printf("[DEBUG] found the following hosts: %+v", hosts)
	if err != nil {
		return err
	}

	if err = d.Set("hosts", hostsToMapList(hosts)); err != nil {
		return err
	}
	if err = d.Set("master", pool[0].Master); err != nil {
		return err
	}

	d.SetId(pool[0].Master)
	return nil
}

func hostsToMapList(hosts []client.Host) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(hosts))
	for _, host := range hosts {
		hostMap := map[string]interface{}{
			"id":           host.Id,
			"name_label":   host.NameLabel,
			"pool_id":      host.Pool,
			"tags":         host.Tags,
			"memory":       host.Memory.Size,
			"memory_usage": host.Memory.Usage,
			"cpus":         hostCpuInfoToMapList(host),
		}
		result = append(result, hostMap)
	}

	return result
}
