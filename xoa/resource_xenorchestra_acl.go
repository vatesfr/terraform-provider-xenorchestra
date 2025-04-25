package xoa

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

var validActionOptions = []string{
	"admin",
	"operator",
	"viewer",
}

func resourceAcl() *schema.Resource {
	return &schema.Resource{
		Create: resourceAclCreate,
		Read:   resourceAclRead,
		Delete: resourceAclDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"subject": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The uuid of the user account that the acl will apply to.",
			},
			"object": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The id of the object that will be able to be used by the subject.",
			},
			"action": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(validActionOptions, false),
				Description:  "Must be one of admin, operator, viewer. See the [Xen orchestra docs](https://xen-orchestra.com/docs/acls.html) on ACLs for more details.",
			},
		},
	}
}

func resourceAclCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	acl, err := c.V1Client().CreateAcl(client.Acl{
		Subject: d.Get("subject").(string),
		Object:  d.Get("object").(string),
		Action:  d.Get("action").(string),
	})
	if err != nil {
		return err
	}
	return aclToData(acl, d)
}

func resourceAclRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	acl, err := c.V1Client().GetAcl(client.Acl{
		Id: d.Id(),
	})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	return aclToData(acl, d)
}

func resourceAclDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	err := c.V1Client().DeleteAcl(client.Acl{
		Id: d.Id(),
	})

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func aclToData(acl *client.Acl, d *schema.ResourceData) error {
	d.SetId(acl.Id)
	if err := d.Set("subject", acl.Subject); err != nil {
		return err
	}
	if err := d.Set("object", acl.Object); err != nil {
		return err
	}
	if err := d.Set("action", acl.Action); err != nil {
		return err
	}
	return nil
}
