package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var netDefaultDesc string = "Created with Xen Orchestra"
var validBondModes []string = []string{"balance-slb", "active-backup", "lacp"}

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
				ConflictsWith: []string{
					"pif_ids",
				},
				ForceNew:    true,
				Description: "The pif (uuid) that should be used for this network.",
			},
			"pif_ids": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ConflictsWith: []string{
					"pif_id",
					"nbd",
				},
				RequiredWith: []string{
					"bond_mode",
				},
				ForceNew:    true,
				Description: "The pifs (uuid) that should be used for this network.",
			},
			"bond_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The bond mode that should be used for this network.",
				ValidateFunc: validation.StringInSlice(validBondModes, false),
			},
			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  -1,
				RequiredWith: []string{
					"pif_id",
				},
				ForceNew:    true,
				Description: "The vlan to use for the network. Defaults to `-1` meaning no VLAN.",
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

	pifsReq := []string{}
	for _, pif := range d.Get("pif_ids").([]interface{}) {
		pifsReq = append(pifsReq, pif.(string))
	}
	// The Xen Orchestra API treats the VLAN default as 0, however,
	// the xapi will return a -1 in those cases. Terraform must treat
	vlan := 0
	if vlanParam, _ := d.Get("vlan").(int); vlanParam != -1 {
		vlan = vlanParam
	}
	network, err := c.CreateNetwork(client.CreateNetworkRequest{
		BondMode:        d.Get("bond_mode").(string),
		Automatic:       d.Get("automatic").(bool),
		DefaultIsLocked: d.Get("default_is_locked").(bool),
		Name:            d.Get("name_label").(string),
		Description:     d.Get("name_description").(string),
		Pool:            d.Get("pool_id").(string),
		Mtu:             d.Get("mtu").(int),
		Nbd:             d.Get("nbd").(bool),
		Vlan:            vlan,
		PIF:             d.Get("pif_id").(string),
		PIFs:            pifsReq,
	})
	if err != nil {
		return err
	}
	var pif *client.PIF
	if len(network.PIFs) > 0 {
		pifs, err := c.GetPIF(client.PIF{Id: network.PIFs[0]})
		if err != nil {
			return err
		}
		pif = &pifs[0]
	}
	return networkToData(network, pif, d)
}

// func getVlanForNetwork(c client.XOClient, net *client.Network) (int, error) {
// 	if len(net.PIFs) > 0 {
// 		pifs, err := c.GetPIF(client.PIF{Id: net.PIFs[0]})
// 		if err != nil {
// 			return -1, err
// 		}

// 		if len(pifs) != 1 {
// 			return -1, errors.New("expected to find single PIF")
// 		}
// 		return pifs[0].Vlan, nil
// 	}
// 	return 0, nil
// }

func resourceNetworkRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)
	network, err := c.GetNetwork(
		client.Network{Id: d.Id()})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	var pif *client.PIF
	if len(network.PIFs) > 0 {
		pifs, err := c.GetPIF(client.PIF{Id: network.PIFs[0]})
		if err != nil {
			return err
		}
		pif = &pifs[0]
	}
	return networkToData(network, pif, d)
}

func resourceNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	netUpdateReq := client.UpdateNetworkRequest{
		Id:              d.Id(),
		Automatic:       d.Get("automatic").(bool),
		DefaultIsLocked: d.Get("default_is_locked").(bool),
		Nbd:             d.Get("nbd").(bool),
	}
	if d.HasChange("name_label") {
		nameLabel := d.Get("name_label").(string)
		netUpdateReq.NameLabel = &nameLabel
	}
	if d.HasChange("name_description") {
		nameDescription := d.Get("name_description").(string)
		netUpdateReq.NameDescription = &nameDescription
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

func networkToData(network *client.Network, pif *client.PIF, d *schema.ResourceData) error {
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
	if pif != nil {
		if err := d.Set("vlan", pif.Vlan); err != nil {
			return err
		}
	}
	return nil
}
