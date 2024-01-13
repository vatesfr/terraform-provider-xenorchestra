package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXoaVDI() *schema.Resource {
	return &schema.Resource{
		Description: `Provides information about a VDI (virtual disk image).

**Note:** If there are multiple VDIs that match terraform will fail.
Ensure that your name_label, pool_id and tags identify a unique VDI.`,
		Read: dataSourceVDIRead,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the VDI to look up.",
				Required:    true,
			},
			"parent": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The ID of the parent VDI if one exists. An example of when a VDI will have a parent is when it was created from a VM fast clone.",
				Computed:    true,
			},
			"pool_id": &schema.Schema{
				Description: "The ID of the pool the VDI belongs to. This is useful if you have a VDI with the same name on different pools.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags": resourceTags(),
		},
	}
}

func dataSourceVDIRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)
	tags := d.Get("tags").(*schema.Set).List()

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
	d.Set("parent", vdi.Parent)
	return nil
}
