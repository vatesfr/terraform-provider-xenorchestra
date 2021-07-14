package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXoaHost() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceHostRead,
		Schema: resourceHostSchema(),
	}
}

func resourceHost() *schema.Resource {
	hostSchema := resourceHostSchema()
	// This is needed by the hosts data source but
	// will cause problems for the host data source.
	// Add this map key at runtime to allow for code reuse.
	hostSchema["id"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}
	return &schema.Resource{
		Schema: hostSchema,
	}
}

func resourceHostSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name_label": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"pool_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"cpus": &schema.Schema{
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeInt},
		},
		"memory": &schema.Schema{
			Type:     schema.TypeInt,
			Computed: true,
		},
		"memory_usage": &schema.Schema{
			Type:     schema.TypeInt,
			Computed: true,
		},
		"tags": resourceTags(),
	}
}

func dataSourceHostRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)
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
	d.Set("pool_id", hosts[0].Pool)
	d.Set("memory", hosts[0].Memory.Size)
	d.Set("memory_usage", hosts[0].Memory.Usage)
	d.Set("cpus", hostCpuInfoToMapList(hosts[0]))
	d.Set("tags", hosts[0].Tags)
	return nil
}

func hostCpuInfoToMapList(host client.Host) map[string]int {
	return map[string]int{
		"sockets": int(host.Cpus.Sockets),
		"cores":   int(host.Cpus.Cores),
	}
}
