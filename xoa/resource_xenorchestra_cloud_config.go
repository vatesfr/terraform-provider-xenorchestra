package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform/helper/schema"
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
	c, err := client.NewClient()
	if err != nil {
		return err
	}

	config, err := c.CreateCloudConfig(d.Get("name").(string), d.Get("template").(string))
	if err != nil {
		return err
	}
	d.SetId(config.Id)
	return nil
}

func resourceCloudConfigRead(d *schema.ResourceData, m interface{}) error {
	c, err := client.NewClient()
	if err != nil {
		return err
	}

	config, err := c.GetCloudConfig(d.Id())
	if err != nil {
		return err
	}

	if config == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", config.Name)
	d.Set("template", config.Template)
	return nil
}

func resourceCloudConfigDelete(d *schema.ResourceData, m interface{}) error {
	c, err := client.NewClient()
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

	c, err := client.NewClient()
	if err != nil {
		return nil, err
	}

	config, err := c.GetCloudConfig(d.Id())

	if err != nil {
		return nil, err
	}
	d.Set("name", config.Name)
	d.Set("template", config.Template)
	return []*schema.ResourceData{d}, nil
}
