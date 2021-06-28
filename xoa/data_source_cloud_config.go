package xoa

import (
	"errors"
	"fmt"
	"log"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXoaCloudConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudConfigRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"template": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCloudConfigRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	name := d.Get("name").(string)

	cloudConfigs, err := c.GetCloudConfigByName(name)

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
