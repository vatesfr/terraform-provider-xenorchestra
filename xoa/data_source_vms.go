package xoa

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/terraform-provider-xenorchestra/xoa/internal"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaVms() *schema.Resource {

	return &schema.Resource{
		Description: "Use this data source to filter Xenorchestra VMs by certain criteria (pool_id, power_state or host) for use in other resources.",
		ReadContext: dataSourceVmsReadContext,
		Schema: map[string]*schema.Schema{
			"vms": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        resourceVm(),
				Description: "A list of information for all vms found in this pool.",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The ID of the pool the VM belongs to.",
				Required:    true,
			},
			"host": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"power_state": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The power state of the vms. (Running, Halted)",
				Optional:    true,
			},
		},
	}
}

func dataSourceVmsReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)
	searchVm := client.Vm{
		PowerState: d.Get("power_state").(string),
		Host:       d.Get("host").(string),
		PoolId:     d.Get("pool_id").(string),
	}

	vms, err := c.GetVms(searchVm)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("vms", vmToMapList(ctx, vms)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(internal.Strings([]string{searchVm.PowerState, searchVm.PoolId, searchVm.Host}))
	return nil

}

func vmToMapList(ctx context.Context, vms []client.Vm) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(vms))
	for _, vm := range vms {
		tflog.Debug(ctx, "Found IPs", map[string]interface{}{"addresses": vm.Addresses})
		var ipv4 []string
		var ipv6 []string
		for key, address := range vm.Addresses {
			if strings.Contains(key, "ipv4") {
				ipv4 = append(ipv4, address)
			} else if strings.Contains(key, "ipv6") {
				ipv6 = append(ipv6, address)
			}
		}
		tflog.Debug(ctx, "VBD info", map[string]interface{}{"vbds": vm.VBDs, "nameLabel": vm.NameLabel, "id": vm.Id})
		hostMap := map[string]interface{}{
			"id":                   vm.Id,
			"name_label":           vm.NameLabel,
			"cpus":                 vm.CPUs.Number,
			"cloud_config":         vm.CloudConfig,
			"cloud_network_config": vm.CloudNetworkConfig,
			"tags":                 vm.Tags,
			"memory_max":           vm.Memory.Static[1],
			"memory_min":           vm.Memory.Static[0],
			"affinity_host":        vm.AffinityHost,
			"template":             vm.Template,
			"high_availability":    vm.HA,
			"ipv4_addresses":       ipv4,
			"ipv6_addresses":       ipv6,
			"power_state":          vm.PowerState,
			"host":                 vm.Host,
			"auto_poweron":         vm.AutoPoweron,
			"name_description":     vm.NameDescription,
		}
		if vm.ResourceSet != nil {
			hostMap["resource_set"] = vm.ResourceSet.Id
		}
		result = append(result, hostMap)
	}

	return result
}
