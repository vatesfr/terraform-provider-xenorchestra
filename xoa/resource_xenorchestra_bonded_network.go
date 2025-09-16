package xoa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

var validBondModes []string = []string{"balance-slb", "active-backup", "lacp"}

func resourceXoaBondedNetwork() *schema.Resource {
	return &schema.Resource{
		Description:   "A resource for managing Bonded Xen Orchestra networks. See the XCP-ng [networking docs](https://xcp-ng.org/docs/networking.html) for more details.",
		CreateContext: resourceBondedNetworkCreateContext,
		DeleteContext: resourceBondedNetworkDeleteContext,
		ReadContext:   resourceBondedNetworkReadContext,
		UpdateContext: resourceBondedNetworkUpdateContext,
		Importer: &schema.ResourceImporter{
			State: resourceBondedNetworkImport,
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
			"pif_ids": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "The PIFs (uuid) that should be used for this network.",
			},
			"bond_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
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

func resourceBondedNetworkCreateContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.FromErr(err)
	}
	if len(network.PIFs) < 1 {
		return diag.FromErr(fmt.Errorf("network should contain more than one PIF after creation"))
	}
	if err := bondedNetworkToData(network, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceBondedNetworkReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)
	network, err := c.GetNetwork(
		client.Network{Id: d.Id()})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if len(network.PIFs) < 1 {
		return diag.FromErr(fmt.Errorf("network should contain more than one PIF"))
	}

	if err := bondedNetworkToData(network, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceBondedNetworkUpdateContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	netUpdateReq := client.UpdateNetworkRequest{
		Id:        d.Id(),
		Automatic: d.Get("automatic").(bool),
	}
	if d.HasChange("name_label") {
		nameLabel := d.Get("name_label").(string)
		netUpdateReq.NameLabel = &nameLabel
	}
	if d.HasChange("name_description") {
		nameDescription := d.Get("name_description").(string)
		netUpdateReq.NameDescription = &nameDescription
	}
	if d.HasChange("default_is_locked") {
		defaultIsLocked := d.Get("default_is_locked").(bool)
		netUpdateReq.DefaultIsLocked = &defaultIsLocked
	}
	_, err := c.UpdateNetwork(netUpdateReq)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceBondedNetworkReadContext(ctx, d, m)
}

func resourceBondedNetworkDeleteContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	err := c.DeleteNetwork(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

// Custom importer to populate pif_ids from BondSlaves of the network's main PIF
func resourceBondedNetworkImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(client.XOClient)
	network, err := c.GetNetwork(client.Network{Id: d.Id()})
	if err != nil {
		return nil, err
	}

	if len(network.PIFs) < 1 {
		return nil, fmt.Errorf("network should contain more than one PIF")
	}

	// Get the bonded pifs and bond mode from the master pif
	for _, pifID := range network.PIFs {
		bondPifs, err := c.GetPIF(client.PIF{Id: pifID})
		if err != nil {
			return nil, err
		}
		if len(bondPifs) < 1 {
			return nil, fmt.Errorf("no PIF returned for ID: %s", pifID)
		}
		if bondPifs[0].IsBondMaster {
			if err := d.Set("pif_ids", bondPifs[0].BondSlaves); err != nil {
				return nil, err
			}
			bond, err := c.GetBond(client.Bond{Master: bondPifs[0].Id})
			if err != nil {
				return nil, err
			}
			d.Set("bond_mode", bond.Mode)
			break
		}
	}

	if err := bondedNetworkToData(network, d); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
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
