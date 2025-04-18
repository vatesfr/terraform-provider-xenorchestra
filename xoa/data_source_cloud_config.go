package xoa

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

func dataSourceXoaCloudConfig() *schema.Resource {
	return &schema.Resource{
		Description: `Provides information about cloud config.

**NOTE:** If there are multiple cloud configs with the same name Terraform will fail. Ensure that your names are unique when using the data source.`,
		Read: dataSourceCloudConfigRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the cloud config you want to look up.",
			},
			"template": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The contents of the cloud-config.",
			},
		},
	}
}

func dataSourceCloudConfigRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)

	name := d.Get("name").(string)

	cloudConfigs, err := c.V1Client().GetCloudConfigByName(name)

	if err != nil {
		return err
	}

	l := len(cloudConfigs)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` cloud configs with name `%s`. Cloud configs must be uniquely named to use this data source", l, name))
	}

	cloudConfig := cloudConfigs[0]

	log.Printf("[DEBUG] Found cloud config with %+v", cloudConfig)
	d.SetId(cloudConfig.Id)
	if err := d.Set("template", cloudConfig.Template); err != nil {
		return err
	}
	return nil
}
