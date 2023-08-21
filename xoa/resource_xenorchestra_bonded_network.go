package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validBondModes []string = []string{"balance-slb", "active-backup", "lacp"}

func resourceXoaBondedNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceBondedNetworkCreate,
		Delete: resourceBondedNetworkDelete,
		Read:   resourceBondedNetworkRead,
		Update: resourceBondedNetworkUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			// Verify if network.set applies for bonded networks unconditionally or if it
			// only works with a subset of the parameters

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
			"pif_ids": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
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
		},
	}
}

func resourceBondedNetworkCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	pifsReq := []string{}
	for _, pif := range d.Get("pif_ids").([]interface{}) {
		pifsReq = append(pifsReq, pif.(string))
	}
	network, err := c.CreateBondedNetwork(client.CreateBondedNetworkRequest{
		Automatic:       d.Get("automatic").(bool),
		DefaultIsLocked: d.Get("default_is_locked").(bool),
		BondMode:        d.Get("bond_mode").(string),
		Name:            d.Get("name_label").(string),
		Description:     d.Get("name_description").(string),
		Pool:            d.Get("pool_id").(string),
		Mtu:             d.Get("mtu").(int),
		PIFs:            pifsReq,
	})
	if err != nil {
		return err
	}
	if len(network.PIFs) < 1 {
		return errors.New("network should contain more than one PIF")
	}
	fmt.Printf("[WARNING] attempting to set pif_ids\n")
	if err := d.Set("pif_ids", pifsReq); err != nil {
		return errors.New("failed to set pif_ids attribute.")
	}
	return bondedNetworkToData(network, d)
}

func resourceBondedNetworkRead(d *schema.ResourceData, m interface{}) error {
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

	if len(network.PIFs) < 1 {
		return errors.New("network should contain more than one PIF")
	}
	return bondedNetworkToData(network, d)
}

func resourceBondedNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	netUpdateReq := client.UpdateNetworkRequest{
		Id:              d.Id(),
		Automatic:       d.Get("automatic").(bool),
		DefaultIsLocked: d.Get("default_is_locked").(bool),
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
	return resourceBondedNetworkRead(d, m)
}

func resourceBondedNetworkDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	err := c.DeleteNetwork(d.Id())

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func bondedNetworkToData(network *client.Network, d *schema.ResourceData) error {
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

	if err := d.Set("automatic", network.Automatic); err != nil {
		return err
	}
	if err := d.Set("default_is_locked", network.DefaultIsLocked); err != nil {
		return err
	}
	return nil
}
