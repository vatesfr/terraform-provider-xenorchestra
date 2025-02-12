package xoa

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

var validTypes = []string{
	"raw",
	// TODO(ddelnano): support vhd when the use case is better understood
	// "vhd",
}

func resourceVDIRecord() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a Xen Orchestra vdi resource.",
		Create:      resourceVDICreate,
		Read:        resourceVDIRead,
		Update:      resourceVDIUpdate,
		Delete:      resourceVDIDelete,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name label of the VDI",
				Required:    true,
			},
			"sr_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The id of the storage repository the VDI should be created in. Make sure the storage repository supports the file you are uploading! For example, ISOs should only be uploaded to ISO storage repositories.",
				Required:    true,
				ForceNew:    true,
			},
			"filepath": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The file path to the ISO or vdi image that should be uploaded when the VDI is created.",
				ForceNew:    true,
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Only `raw` uploads are supported today, but vhd support may be added in the future.",
				ValidateFunc: validation.StringInSlice(validTypes, false),
			},
		},
	}
}

func resourceVDICreate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	vdi, err := c.CreateVDI(client.CreateVDIReq{
		NameLabel: d.Get("name_label").(string),
		SRId:      d.Get("sr_id").(string),
		Filepath:  d.Get("filepath").(string),
		Type:      d.Get("type").(string),
	})
	if err != nil {
		return err
	}
	d.SetId(vdi.VDIId)
	return vdiToData(vdi, d)
}

func resourceVDIRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	vdi, err := c.GetVDI(client.VDI{VDIId: d.Id()})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	return vdiToData(vdi, d)
}

func resourceVDIUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	err := c.UpdateVDI(client.Disk{
		VDI: client.VDI{
			VDIId:     d.Id(),
			NameLabel: d.Get("name_label").(string),
		},
	})
	if err != nil {
		return err
	}

	vdi, err := c.GetVDI(client.VDI{VDIId: d.Id()})
	if err != nil {
		return err
	}

	return vdiToData(vdi, d)
}

func resourceVDIDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	err := c.DeleteVDI(d.Id())

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func vdiToData(vdi client.VDI, d *schema.ResourceData) error {
	d.SetId(vdi.VDIId)
	keys := map[string]string{
		"name_label": vdi.NameLabel,
		"sr_id":      vdi.SrId,
	}
	for k, v := range keys {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	if err := d.Set("filepath", d.Get("filepath")); err != nil {
		return err
	}
	if err := d.Set("type", d.Get("type")); err != nil {
		return err
	}
	return nil
}
