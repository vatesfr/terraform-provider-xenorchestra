package xoa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

var validActionOptions = []string{
	"admin",
	"operator",
	"viewer",
}

func resourceAcl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAclCreateContext,
		ReadContext:   resourceAclReadContext,
		DeleteContext: resourceAclDeleteContext,
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

func resourceAclCreateContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	acl, err := c.CreateAcl(client.Acl{
		Subject: d.Get("subject").(string),
		Object:  d.Get("object").(string),
		Action:  d.Get("action").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if err := aclToData(acl, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAclReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	acl, err := c.GetAcl(client.Acl{
		Id: d.Id(),
	})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if err := aclToData(acl, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceAclDeleteContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	err := c.DeleteAcl(client.Acl{
		Id: d.Id(),
	})

	if err != nil {
		return diag.FromErr(err)
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
