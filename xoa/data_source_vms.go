package xoa

import (
	"log"
	"strings"

	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xo-sdk-go/client"
)

func dataSourceXoaVms() *schema.Resource {

	return &schema.Resource{
		Read: dataSourceVmsRead,
		Schema: map[string]*schema.Schema{
			"vms": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     resourceVm(),
			},
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"host": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"power_state": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceVmsRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)
	searchVm := client.Vm{
		PowerState: d.Get("power_state").(string),
		Host:       d.Get("host").(string),
		PoolId:     d.Get("pool_id").(string),
	}

	vms, err := c.GetVms(searchVm)
	if err != nil {
		return err
	}

	if err = d.Set("vms", vmToMapList(vms)); err != nil {
		return err
	}
	d.SetId(internal.Strings([]string{searchVm.PowerState, searchVm.PoolId, searchVm.Host}))
	return nil

}

func vmToMapList(vms []client.Vm) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(vms))
	for _, vm := range vms {
		log.Printf("[DEBUG] IPS %s\n", vm.Addresses)
		var ipv4 []string
		var ipv6 []string
		for key, address := range vm.Addresses {
			if strings.Contains(key, "ipv4") {
				ipv4 = append(ipv4, address)
			} else if strings.Contains(key, "ipv6") {
				ipv6 = append(ipv6, address)
			}
		}
		log.Printf("[DEBUG] VBD on %s (%s) %s\n", vm.VBDs, vm.NameLabel, vm.Id)
		hostMap := map[string]interface{}{
			"id":                   vm.Id,
			"name_label":           vm.NameLabel,
			"cpus":                 vm.CPUs.Number,
			"cloud_config":         vm.CloudConfig,
			"cloud_network_config": vm.CloudNetworkConfig,
			"tags":                 vm.Tags,
			"memory_max":           vm.Memory.Static[1],
			"affinity_host":        vm.AffinityHost,
			"template":             vm.Template,
			"wait_for_ip":          vm.WaitForIps,
			"high_availability":    vm.HA,
			"resource_set":         vm.ResourceSet,
			"ipv4_addresses":       ipv4,
			"ipv6_addresses":       ipv6,
			"power_state":          vm.PowerState,
			"host":                 vm.Host,
			"auto_poweron":         vm.AutoPoweron,
			"name_description":     vm.NameDescription,
		}
		result = append(result, hostMap)
	}

	return result
}
