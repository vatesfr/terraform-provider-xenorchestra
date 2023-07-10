package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
)

var netDefaultDesc string = "Created with Xen Orchestra"

func resourceXoaNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkCreate,
		Delete: resourceNetworkDelete,
		Read:   resourceNetworkRead,
		Update: resourceNetworkUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"automatic": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"default_is_locked": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name_description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  netDefaultDesc,
			},
			// "pif_id": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// },
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mtu": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1500,
				ForceNew: true,
			},
			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"nbd": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceNetworkCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	network, err := c.CreateNetwork(client.Network{
		NameLabel:       d.Get("name_label").(string),
		NameDescription: d.Get("name_description").(string),
		PoolId:          d.Get("pool_id").(string),
		MTU:             d.Get("mtu").(int),
		Nbd:             d.Get("nbd").(bool),
	})
	if err != nil {
		return err
	}
	return networkToData(network, d)
}

func resourceNetworkRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)

	net, err := c.GetNetwork(client.Network{
		NameLabel: nameLabel,
		PoolId:    poolId,
	})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}
	return networkToData(net, d)
}

func resourceNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	attrUpdates := map[string]string{
		"automatic":         "",
		"default_is_locked": "defaultIsLocked",
		"nbd":               "",
		"name_label":        "",
		"name_description":  "",
	}
	params := map[string]interface{}{
		"id": d.Id(),
	}
	for tfAttr, xoAttr := range attrUpdates {
		if d.HasChange(tfAttr) {
			attr := tfAttr
			if xoAttr != "" {
				attr = xoAttr
			}
			params[attr] = d.Get(tfAttr)
		}
	}
	var netUpdateReq client.Network
	if err := mapstructure.Decode(params, &netUpdateReq); err != nil {
		return err
	}
	_, err := c.UpdateNetwork(netUpdateReq)
	if err != nil {
		return err
	}
	return resourceNetworkRead(d, m)
}

func resourceNetworkDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	err := c.DeleteNetwork(d.Id())

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func networkToData(network *client.Network, d *schema.ResourceData) error {
	d.SetId(network.Id)
	if err := d.Set("name_label", network.NameLabel); err != nil {
		return err
	}
	if err := d.Set("name_description", network.NameDescription); err != nil {
		return err
	}
	if err := d.Set("pool_id", network.PoolId); err != nil {
		return err
	}
	if err := d.Set("mtu", network.MTU); err != nil {
		return err
	}
	if err := d.Set("nbd", network.Nbd); err != nil {
		return err
	}
	if err := d.Set("automatic", network.Automatic); err != nil {
		return err
	}
	if err := d.Set("default_is_locked", network.DefaultIsLocked); err != nil {
		return err
	}
	return nil
}
