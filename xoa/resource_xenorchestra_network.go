package xoa

import (
	"errors"

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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "This argument controls whether the network should enforce VIF locking. This defaults to `false` which means that no filtering rules are applied.",
			},
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name label of the network.",
			},
			"name_description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  netDefaultDesc,
			},
			"pif_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				RequiredWith: []string{
					"vlan",
				},
				ForceNew:    true,
				Description: "The pif (uuid) that should be used for this network.",
			},
			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				RequiredWith: []string{
					"pif_id",
				},
				ForceNew:    true,
				Description: "The vlan to use for the network. Defaults to `0` meaning no VLAN.",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The pool id that this network should belong to.",
			},
			"mtu": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1500,
				ForceNew:    true,
				Description: "The MTU of the network. Defaults to `1500` if unspecified.",
			},
			"nbd": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the network should use a network block device. Defaults to `false` if unspecified.",
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
	}, d.Get("vlan").(int), d.Get("pif_id").(string))
	if err != nil {
		return err
	}
	vlan := 0
	if len(network.PIFs) > 0 {
		pifs, err := c.GetPIF(client.PIF{Id: network.PIFs[0]})
		if err != nil {
			return err
		}

		if len(pifs) != 1 {
			return errors.New("expected to find single PIF")
		}
		vlan = pifs[0].Vlan
	}
	return networkToData(network, vlan, d)
}

func resourceNetworkRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	vlan := 0
	net, err := c.GetNetwork(
		client.Network{Id: d.Id()})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}
	if len(net.PIFs) > 0 {
		pifs, err := c.GetPIF(client.PIF{Id: net.PIFs[0]})
		if err != nil {
			return err
		}

		if len(pifs) != 1 {
			return errors.New("expected to find single PIF")
		}
		vlan = pifs[0].Vlan
	}
	return networkToData(net, vlan, d)
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

func networkToData(network *client.Network, vlan int, d *schema.ResourceData) error {
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
	if err := d.Set("vlan", vlan); err != nil {
		return err
	}
	return nil
}
