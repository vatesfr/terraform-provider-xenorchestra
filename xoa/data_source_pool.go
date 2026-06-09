package xoa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePoolReadContext,
		Description: "Provides information about a pool.",
		Schema:      resourcePoolSchema(),
	}
}

func resourcePoolSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name_label": &schema.Schema{
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name_label of the pool to look up.",
		},
		"description": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The description of the pool.",
		},
		"master": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The id of the primary instance in the pool.",
		},
		"id": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The id of the pool.",
		},
		"default_sr": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The default storage repository for the pool.",
		},
		"cpus": &schema.Schema{
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "CPU information about the pool. " + cpusDesc,
		},
	}
}

func poolCpuInfoToMap(cpus client.CpuInfo) map[string]interface{} {
	return map[string]interface{}{
		"cores":   fmt.Sprintf("%d", cpus.Cores),
		"sockets": fmt.Sprintf("%d", cpus.Sockets),
	}
}

func poolToMap(pool client.Pool) map[string]interface{} {
	return map[string]interface{}{
		"id":          pool.Id,
		"name_label":  pool.NameLabel,
		"description": pool.Description,
		"master":      pool.Master,
		"cpus":        poolCpuInfoToMap(pool.Cpus),
		"default_sr":  pool.DefaultSR,
	}
}

func dataSourcePoolReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)

	pools, err := c.GetPoolByName(nameLabel)

	if err != nil {
		return diag.FromErr(err)
	}

	l := len(pools)
	if l != 1 {
		return diag.FromErr(fmt.Errorf("found `%d` pools with name `%s`. Pools must be uniquely named to use this data source", l, nameLabel))
	}

	pool := pools[0]

	tflog.Debug(ctx, "Found pool", map[string]interface{}{
		"pool": pool,
	})

	d.SetId(pool.Id)
	for k, v := range poolToMap(pool) {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
