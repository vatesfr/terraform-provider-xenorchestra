package xoa

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

func resourceCloudConfigRecord() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a Xen Orchestra cloud config resource.",
		Create:      resourceCloudConfigCreate,
		Read:        resourceCloudConfigRead,
		Delete:      resourceCloudConfigDelete,
		Importer: &schema.ResourceImporter{
			State: CloudConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"template": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The cloud init config. See the cloud init docs for more [information](https://cloudinit.readthedocs.io/en/latest/topics/examples.html).",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the cloud config.",
			},
		},
	}
}

func resourceCloudConfigCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	cloud_config, err := c.V1Client().CreateCloudConfig(d.Get("name").(string), d.Get("template").(string))
	if err != nil {
		return err
	}
	d.SetId(cloud_config.Id)
	return nil
}

func resourceCloudConfigRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	cloud_config, err := c.V1Client().GetCloudConfig(d.Id())
	if err != nil {
		return err
	}

	if cloud_config == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", cloud_config.Name)
	d.Set("template", cloud_config.Template)
	return nil
}

func resourceCloudConfigDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	err := c.V1Client().DeleteCloudConfig(d.Id())

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func CloudConfigImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	c := m.(*v2.XOClient)

	cloud_config, err := c.V1Client().GetCloudConfig(d.Id())

	if err != nil {
		return nil, err
	}
	d.Set("name", cloud_config.Name)
	d.Set("template", cloud_config.Template)
	return []*schema.ResourceData{d}, nil
}
