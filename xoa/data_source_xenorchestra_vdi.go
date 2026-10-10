package xoa

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/terraform-provider-xenorchestra/xoa/internal"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaVDI() *schema.Resource {
	return &schema.Resource{
		Description: `Provides information about a VDI (virtual disk image).

**Note:** If there are multiple VDIs that match terraform will fail.
Ensure that your name_label, pool_id and tags identify a unique VDI.`,
		Read: dataSourceVDIRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:         schema.TypeString,
				Description:  "The ID of the VDI.",
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"name_label"},
			},
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the VDI to look up.",
				Optional:    true,
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

	tags := d.Get("tags").(*schema.Set).List()
	filter := internal.VDIFilter{
		Id:        d.Get("id").(string),
		NameLabel: d.Get("name_label").(string),
		PoolId:    d.Get("pool_id").(string),
		Tags:      tagsFromInterfaceSlice(tags),
	}

	allVdis, err := c.GetAllVDIs()

	if err != nil {
		return err
	}

	vdis := internal.FilterVDIs(allVdis, filter)

	l := len(vdis)
	if l == 0 {
		return fmt.Errorf("could not find a VDI that matches %+v", filter)
	}

	if l > 1 {
		return fmt.Errorf("found `%d` VDIs that match %+v. VDIs must be uniquely named to use this data source", l, vdis)
	}

	vdi := vdis[0]

	d.SetId(vdi.VDIId)
	d.Set("name_label", vdi.NameLabel)
	d.Set("pool_id", vdi.PoolId)
	d.Set("tags", vdi.Tags)
	d.Set("parent", vdi.Parent)
	return nil
}
