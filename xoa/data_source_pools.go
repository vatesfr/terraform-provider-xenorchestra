package xoa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vatesfr/terraform-provider-xenorchestra/xoa/internal"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaPools() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to filter Xenorchestra pools by certain criteria (tags) for use in other resources.",
		ReadContext: dataSourcePoolsReadContext,
		Schema: map[string]*schema.Schema{
			"pools": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        resourcePool(),
				Description: "The resulting pools after applying the argument filtering.",
			},
			"tags": resourceTags(),
			"sort_by": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"id", "name_label"}, false),
				Description:  "The pool field to sort the results by (id and name_label are supported).",
			},
			"sort_order": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
				Default:      "asc",
				Description:  "Valid options are `asc` or `desc` and sort order is applied to `sort_by` argument.",
			},
		},
	}
}

func dataSourcePoolsReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	// Get filter parameters
	tags := d.Get("tags").(*schema.Set).List()
	tagStrings := make([]string, 0, len(tags))
	for _, t := range tags {
		tagStrings = append(tagStrings, t.(string))
	}

	poolIds, err := c.GetObjectsWithTags(tagStrings)
	if err != nil {
		return diag.FromErr(err)
	}

	// If tags are specified, GetObjectsWithTags returns pools with ALL matching tags
	// If no tags are specified, GetObjectsWithTags returns an error in some SDK versions
	// So we fall back to GetPools with empty filter for "all pools"
	var pools []client.Pool
	if len(tagStrings) > 0 {
		// We got pool IDs from GetObjectsWithTags, now get their details
		pools = make([]client.Pool, 0, len(poolIds))
		for _, obj := range poolIds {
			if obj.Type == "pool" {
				p, err := c.GetPools(client.Pool{Id: obj.Id})
				if err != nil {
					return diag.FromErr(err)
				}
				if len(p) > 0 {
					pools = append(pools, p[0])
				}
			}
		}
	} else {
		// No tags specified, get all pools
		var err error
		pools, err = c.GetPools(client.Pool{})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Sort if requested
	var sortBy, sortOrder string
	if v := d.Get("sort_by").(string); v != "" {
		sortBy = v
		sortOrder = d.Get("sort_order").(string)
		pools = internal.SortPools(pools, sortBy, sortOrder)
	}

	// Convert to map list for Terraform
	poolList := poolsToMapList(pools)

	if err := d.Set("pools", poolList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(internal.Strings([]string{sortBy, sortOrder}))
	tflog.Debug(ctx, "Found pools", map[string]interface{}{
		"count": len(pools),
	})

	return nil
}

func poolsToMapList(pools []client.Pool) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(pools))
	for _, pool := range pools {
		poolMap := map[string]interface{}{
			"id":          pool.Id,
			"name_label":  pool.NameLabel,
			"description": pool.Description,
			"master":      pool.Master,
			"cpus":        poolCpuInfoToMap(pool.Cpus),
			"default_sr":  pool.DefaultSR,
		}
		result = append(result, poolMap)
	}
	return result
}

func poolCpuInfoToMap(cpus client.CpuInfo) map[string]interface{} {
	return map[string]interface{}{
		"cores":   cpus.Cores,
		"sockets": cpus.Sockets,
	}
}

func resourcePool() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the pool.",
			},
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name label of the pool.",
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
			"cpus": &schema.Schema{
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "CPU information about the pool. " + cpusDesc,
			},
			"default_sr": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The default storage repository for the pool.",
			},
		},
	}
}
