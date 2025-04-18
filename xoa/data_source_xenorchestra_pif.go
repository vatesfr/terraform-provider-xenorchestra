package xoa

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

func dataSourceXoaPIF() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePIFRead,
		Description: `Provides information about a physical network interface (PIF) of a XenServer host specified by the interface name or whether it is the management interface.

**Note:** If there are multiple PIFs that match terraform will fail.
Ensure that your device, vlan, host_id and other arguments identify a unique PIF.`,
		Schema: map[string]*schema.Schema{
			"attached": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If the PIF is attached to the network.",
			},
			"device": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the network device. Examples include eth0, eth1, etc. See `ifconfig` for possible devices.",
			},
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The host the PIF is associated with.",
			},
			"network": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The network the PIF is associated with.",
			},
			"host_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The ID of the host that the PIF belongs to.",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The pool the PIF is associated with.",
			},
			"uuid": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The uuid of the PIF.",
			},
			"vlan": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The VLAN the PIF belongs to.",
			},
		},
	}
}

func dataSourcePIFRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	device := d.Get("device").(string)
	vlan := d.Get("vlan").(int)
	host := d.Get("host_id").(string)

	pifReq := client.PIF{
		Device: device,
		Vlan:   vlan,
		Host:   host,
	}

	pifs, err := c.V1Client().GetPIF(pifReq)

	if err != nil {
		return err
	}

	l := len(pifs)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` pifs with device `%s` and vlan `%d`. PIFs must be uniquely named to use this data source", l, device, vlan))
	}

	pif := pifs[0]

	d.SetId(pif.Id)
	d.Set("uuid", pif.Uuid)
	d.Set("device", pif.Device)
	d.Set("host", pif.Host)
	d.Set("attached", pif.Attached)
	d.Set("pool_id", pif.PoolId)
	d.Set("network", pif.Network)
	d.Set("vlan", pif.Vlan)
	return nil
}
