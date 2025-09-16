package xoa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

var netDefaultDesc string = "Created with Xen Orchestra"

func resourceXoaNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkCreateContext,
		DeleteContext: resourceNetworkDeleteContext,
		ReadContext:   resourceNetworkReadContext,
		UpdateContext: resourceNetworkUpdateContext,
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
			"source_pif_device": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				RequiredWith: []string{
					"vlan",
				},
				ForceNew:    true,
				Description: "The PIF device (eth0, eth1, etc) that will be used as an input during network creation. This parameter is required if a vlan is specified.",
			},
			"vlan": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				RequiredWith: []string{
					"source_pif_device",
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

func resourceNetworkCreateContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	var pifId string
	if sourcePIFDevice := d.Get("source_pif_device").(string); sourcePIFDevice != "" {
		pif, err := getNetworkCreationSourcePIF(c, sourcePIFDevice, d.Get("pool_id").(string))

		if err != nil {
			return diag.FromErr(err)
		}
		pifId = pif.Id
	}

	network, err := c.CreateNetwork(client.CreateNetworkRequest{
		Automatic:       d.Get("automatic").(bool),
		DefaultIsLocked: d.Get("default_is_locked").(bool),
		Name:            d.Get("name_label").(string),
		Description:     d.Get("name_description").(string),
		Pool:            d.Get("pool_id").(string),
		Mtu:             d.Get("mtu").(int),
		Nbd:             d.Get("nbd").(bool),
		Vlan:            d.Get("vlan").(int),
		PIF:             pifId,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	vlan, pifDevice, err := getVlanForNetwork(c, network)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := networkToData(ctx, network, vlan, pifDevice, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// This function returns the PIF specified the given device name on the pool's primary host. In order to create
// networks with a VLAN, a PIF for the given device must be provided. Xen Orchestra uses the primary host's PIF
// for this and so we emulate that behavior.
func getNetworkCreationSourcePIF(c client.XOClient, pifDevice, poolId string) (*client.PIF, error) {
	pools, err := c.GetPools(client.Pool{Id: poolId})
	if err != nil {
		return nil, err
	}

	if len(pools) != 1 {
		return nil, fmt.Errorf("expected to find a single pool, instead found %d", len(pools))
	}

	pool := pools[0]
	pifs, err := c.GetPIF(client.PIF{
		Host:   pool.Master,
		Vlan:   -1,
		Device: pifDevice,
	})

	if err != nil {
		return nil, err
	}

	if len(pifs) != 1 {
		return nil, fmt.Errorf("expected to find a single PIF, instead found %d. %+v", len(pifs), pifs)
	}

	return &pifs[0], nil
}

// Returns the VLAN and device name for the given network.
func getVlanForNetwork(c client.XOClient, net *client.Network) (int, string, error) {
	if len(net.PIFs) > 0 {
		pifs, err := c.GetPIF(client.PIF{Id: net.PIFs[0]})
		if err != nil {
			return -1, "", err
		}

		if len(pifs) != 1 {
			return -1, "", fmt.Errorf("expected to find single PIF")
		}
		return pifs[0].Vlan, pifs[0].Device, nil
	}
	return 0, "", nil
}

func resourceNetworkReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)
	net, err := c.GetNetwork(
		client.Network{Id: d.Id()})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	vlan, pifDevice, err := getVlanForNetwork(c, net)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := networkToData(ctx, net, vlan, pifDevice, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceNetworkUpdateContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	netUpdateReq := client.UpdateNetworkRequest{
		Id:        d.Id(),
		Automatic: d.Get("automatic").(bool),
		Nbd:       d.Get("nbd").(bool),
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
	return resourceNetworkReadContext(ctx, d, m)
}

func resourceNetworkDeleteContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	err := c.DeleteNetwork(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func networkToData(ctx context.Context, network *client.Network, vlan int, pifDevice string, d *schema.ResourceData) error {
	d.SetId(network.Id)
	if err := d.Set("name_label", network.NameLabel); err != nil {
		return err
	}
	tflog.Debug(ctx, "Reading network data")
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
	if err := d.Set("source_pif_device", pifDevice); err != nil {
		return err
	}
	return nil
}
