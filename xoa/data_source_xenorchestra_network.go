package xoa

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xo-sdk-go/client"
)

func dataSourceXoaNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNetworkRead,
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
