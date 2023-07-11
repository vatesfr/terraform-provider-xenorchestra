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
			"bridge": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"pif_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"mtu": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceNetworkCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	var net = &client.Network{}
	networkFromResourceData(net, d)

	net, err := c.CreateNetwork(*net)

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

	resourceDataFromNetwork(d, net)
	return nil
}

func networkFromResourceData(net *client.Network, d *schema.ResourceData) {
	net.Id = d.Id()
	net.Bridge = d.Get("bridge").(string)
	net.NameLabel = d.Get("name_label").(string)
	net.PoolId = d.Get("pool_id").(string)
	net.PifId = d.Get("pif").(string)
	net.Description = d.Get("description").(string)
	net.Mtu = d.Get("mtu").(int)
	net.Vlan = d.Get("vlan").(int)
}

func resourceDataFromNetwork(d *schema.ResourceData, net *client.Network) {
	d.SetId(net.Id)
	d.Set("bridge", net.Bridge)
	d.Set("name_label", net.NameLabel)
	d.Set("pool_id", net.PoolId)
	d.Set("pif", net.PifId)
	d.Set("description", net.Description)
	d.Set("mtu", net.Mtu)
	d.Set("vlan", net.Vlan)
}
