package xoa

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaHost() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceHostRead,
		Schema: resourceHostSchema(),
		Description: `Provides information about a host.

**NOTE:** If there are multiple hosts with the same name
Terraform will fail. Ensure that your names are unique when
using the data source.
		`,
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
		// Description: "The id of the host.",
	}
	return &schema.Resource{
		Schema: hostSchema,
	}
}

var cpusDesc string = "The 'cores' key will contain the number of cpu cores and the 'sockets' key will contain the number of sockets."

func resourceHostSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name_label": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name label of the host.",
		},
		"pool_id": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Id of the pool that the host belongs to.",
		},
		"cpus": &schema.Schema{
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeInt},
			Description: "CPU information about the host. " + cpusDesc,
		},
		"memory": &schema.Schema{
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The memory size of the host.",
		},
		"memory_usage": &schema.Schema{
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The memory usage of the host.",
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
