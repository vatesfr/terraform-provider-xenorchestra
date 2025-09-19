package xoa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaCloudConfig() *schema.Resource {
	return &schema.Resource{
		Description: `Provides information about cloud config.

**NOTE:** If there are multiple cloud configs with the same name Terraform will fail. Ensure that your names are unique when using the data source.`,
		ReadContext: dataSourceCloudConfigReadContext,
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

func dataSourceCloudConfigReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	name := d.Get("name").(string)

	cloudConfigs, err := c.GetCloudConfigByName(name)

	if err != nil {
		return diag.FromErr(err)
	}

	l := len(cloudConfigs)
	if l != 1 {
		return diag.FromErr(fmt.Errorf("found `%d` cloud configs with name `%s`. Cloud configs must be uniquely named to use this data source", l, name))
	}

	cloudConfig := cloudConfigs[0]

	tflog.Debug(ctx, "Found cloud config", map[string]interface{}{
		"cloud_config": cloudConfig,
	})

	d.SetId(cloudConfig.Id)
	if err := d.Set("template", cloudConfig.Template); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
