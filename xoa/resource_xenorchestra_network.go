package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkCreate,
		Read:   resourceNetworkRead,
		Delete: resourceNetworkDelete,
		Schema: map[string]*schema.Schema{
			"bridge": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"name_label": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"pif": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"vlan": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	var net = &client.Network{}
	networkFromResourceData(net, d)
	var vlan = d.Get("vlan").(int)

	net, err := c.CreateNetwork(*net, vlan)

	if err != nil {
		return err
	}

	resourceDataFromNetwork(d, net)
	return nil
}

func resourceNetworkRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)

	net, err := c.GetNetwork(client.Network{
		NameLabel: nameLabel,
		PoolId:    poolId,
	})

	if err != nil {
		return err
	}

	resourceDataFromNetwork(d, net)
	return nil
}

func resourceNetworkDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)

	net, err := c.GetNetwork(client.Network{
		NameLabel: nameLabel,
		PoolId:    poolId,
	})

	if err != nil {
		return err
	}

	err = c.DeleteNetwork(net.Id)

	if err != nil {
		return nil
	}

	d.SetId("")
	return nil
}

func networkFromResourceData(net *client.Network, d *schema.ResourceData) {
	net.Id = d.Id()
	net.Bridge = d.Get("bridge").(string)
	net.NameLabel = d.Get("name_label").(string)
	net.PoolId = d.Get("pool_id").(string)
	net.PifIds = []string{d.Get("pif").(string)}
	net.Description = d.Get("description").(string)
	net.Mtu = d.Get("mtu").(int)
}

func resourceDataFromNetwork(d *schema.ResourceData, net *client.Network) {
	d.SetId(net.Id)
	d.Set("bridge", net.Bridge)
	d.Set("name_label", net.NameLabel)
	d.Set("pool_id", net.PoolId)
	d.Set("pif", net.PifIds[0])
	d.Set("description", net.Description)
	d.Set("mtu", net.Mtu)
}
