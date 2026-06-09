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

	var pools []client.Pool
	var err error
	if len(tagStrings) > 0 {
		pools, err = getPoolsWithTags(c, tagStrings)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// No tags specified, get all pools
		poolMap := make(map[string]client.Pool)
		if err := c.GetAllObjectsOfType(client.Pool{}, &poolMap); err != nil {
			return diag.FromErr(err)
		}
		pools = make([]client.Pool, 0, len(poolMap))
		for _, p := range poolMap {
			pools = append(pools, p)
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
		result = append(result, poolToMap(pool))
	}
	return result
}

func resourcePool() *schema.Resource {
	poolSchema := resourcePoolSchema()
	poolSchema["name_label"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The name label of the pool.",
	}
	return &schema.Resource{
		Schema: poolSchema,
	}
}

func getPoolsWithTags(c client.XOClient, tags []string) ([]client.Pool, error) {
	// GetObjectsWithTags returns a list of object references (type and ID) that match the tags
	poolIds, err := c.GetObjectsWithTags(tags)
	if err != nil {
		return nil, err
	}
	// We got pool IDs from GetObjectsWithTags, now get their details
	pools := make([]client.Pool, 0, len(poolIds))
	for _, obj := range poolIds {
		if obj.Type == "pool" {
			p, err := c.GetPools(client.Pool{Id: obj.Id})
			if err != nil {
				return nil, err
			}
			if len(p) > 0 {
				pools = append(pools, p[0])
			}
		}
	}
	return pools, nil
}
