package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceCloudConfigRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudConfigCreate,
		Read:   resourceCloudConfigRead,
		Delete: resourceCloudConfigDelete,
		Importer: &schema.ResourceImporter{
			State: CloudConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCloudConfigCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)
	if err != nil {
		return err
	}

	cloud_config, err := c.CreateCloudConfig(d.Get("name").(string), d.Get("template").(string))
	if err != nil {
		return err
	}
	d.SetId(cloud_config.Id)
	return nil
}

func resourceCloudConfigRead(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)
	if err != nil {
		return err
	}

	cloud_config, err := c.GetCloudConfig(d.Id())
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
	config := m.(client.Config)
	c, err := client.NewClient(config)
	if err != nil {
		return err
	}

	err = c.DeleteCloudConfig(d.Id())

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func CloudConfigImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {

	config := m.(client.Config)
	c, err := client.NewClient(config)
	if err != nil {
		return nil, err
	}

	cloud_config, err := c.GetCloudConfig(d.Id())

	if err != nil {
		return nil, err
	}
	d.Set("name", cloud_config.Name)
	d.Set("template", cloud_config.Template)
	return []*schema.ResourceData{d}, nil
}
