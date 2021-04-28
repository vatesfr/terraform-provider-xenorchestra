package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceXoaVDI() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVDIRead,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": resourceTags(),
		},
	}
}

func dataSourceVDIRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)
	tags := d.Get("tags").([]interface{})

	vdi := client.VDI{
		NameLabel: nameLabel,
		PoolId:    poolId,
		Tags:      tagsFromInterfaceSlice(tags),
	}

	vdis, err := c.GetVDIs(vdi)

	if err != nil {
		return err
	}

	l := len(vdis)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` VDIs that match %+v. VDIs must be uniquely named to use this data source", l, vdis))
	}

	vdi = vdis[0]

	d.SetId(vdi.VDIId)
	d.Set("name_label", vdi.NameLabel)
	d.Set("pool_id", vdi.PoolId)
	d.Set("tags", vdi.Tags)
	return nil
}
