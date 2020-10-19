package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceXoaPIF() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePIFRead,
		Schema: map[string]*schema.Schema{
			"attached": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"device": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"host": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"network": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"uuid": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func dataSourcePIFRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.Client)

	device := d.Get("device").(string)
	vlan := d.Get("vlan").(int)
	host := d.Get("host_id").(string)

	pifReq := client.PIF{
		Device: device,
		Vlan:   vlan,
		Host:   host,
	}

	pifs, err := c.GetPIF(pifReq)

	if err != nil {
		return err
	}

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
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
