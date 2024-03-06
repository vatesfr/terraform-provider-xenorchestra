package xoa

import (
	"github.com/vatesfr/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXoaNetwork() *schema.Resource {
	return &schema.Resource{
		Description: `Provides information about a network of a Xenserver pool.

**Note:** If there are multiple networks with the same name terraform will fail. Ensure that your network, pool_id and other arguments identify a unique network.`,
		Read: dataSourceNetworkRead,
		Schema: map[string]*schema.Schema{
			"bridge": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the bridge network interface.",
			},
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the network.",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The pool the network is associated with.",
			},
		},
	}
}

func dataSourceNetworkRead(d *schema.ResourceData, m interface{}) error {
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

	d.SetId(net.Id)
	d.Set("bridge", net.Bridge)
	d.Set("name_label", net.NameLabel)
	d.Set("pool_id", net.PoolId)
	return nil
}
