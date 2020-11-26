package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAcl() *schema.Resource {
	return &schema.Resource{
		Create: resourceAclCreate,
		Read:   resourceAclRead,
		Delete: resourceAclDelete,
		Importer: &schema.ResourceImporter{
			State: AclImport,
		},

		Schema: map[string]*schema.Schema{
			"subject": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"object": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAclCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	acl, err := c.CreateAcl(client.Acl{
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
	c := m.(*client.Client)

	acl, err := c.GetAcl(client.Acl{
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
	c := m.(*client.Client)

	err := c.DeleteAcl(client.Acl{
		Id: d.Id(),
	})

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func AclImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	// c := m.(*client.Client)

	// cloud_config, err := c.GetAcl(d.Id())

	// if err != nil {
	// 	return nil, err
	// }
	// d.Set("name", cloud_config.Name)
	// d.Set("template", cloud_config.Template)
	return []*schema.ResourceData{d}, nil
}

func aclToData(acl *client.Acl, d *schema.ResourceData) error {
	d.SetId(acl.Id)
	return nil
}
