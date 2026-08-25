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

func dataSourceXoaSrs() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to filter Xenorchestra storage repositories by certain criteria (name_label, pool_id, host_id, sr_type, tags) for use in other resources.",
		ReadContext: dataSourceSrsReadContext,
		Schema: map[string]*schema.Schema{
			"srs": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        resourceSr(),
				Description: "The resulting storage repositories after applying the argument filtering. `size`, `physical_usage` and `usage` are in bytes.",
			},
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the storage repositories to match.",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Id of the pool the storage repositories exist on.",
			},
			"host_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Id of the host the storage repositories exist on. For host-local storage repositories the SR's `container` is the host itself, so this filters the repositories down to a single host when several hosts in the pool share the same SR name.",
			},
			"sr_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The type of the storage repositories to match (lvm, udev, iso, user, etc).",
			},
			"tags": resourceTags(),
			"sort_by": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"id", "name_label", "size", "physical_usage", "usage"}, false),
				Description:  "The storage repository field to sort the results by (id, name_label, size, physical_usage and usage are supported).",
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

func dataSourceSrsReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	srMap := make(map[string]client.StorageRepository)
	if err := c.GetAllObjectsOfType(client.StorageRepository{}, &srMap); err != nil {
		return diag.FromErr(err)
	}
	srs := make([]client.StorageRepository, 0, len(srMap))
	for _, s := range srMap {
		srs = append(srs, s)
	}

	// The SDK cannot filter by host, but for host-local SRs the SR's
	// `container` is the host that owns it, so filter by container here.
	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)
	hostId := d.Get("host_id").(string)
	srType := d.Get("sr_type").(string)
	tags := tagsFromInterfaceSlice(d.Get("tags").(*schema.Set).List())

	filtered := make([]client.StorageRepository, 0, len(srs))
	for _, s := range srs {
		if nameLabel != "" && s.NameLabel != nameLabel {
			continue
		}
		if poolId != "" && s.PoolId != poolId {
			continue
		}
		if hostId != "" && s.Container != hostId {
			continue
		}
		if srType != "" && s.SRType != srType {
			continue
		}
		if !internal.TagsMatch(tags, s.Tags) {
			continue
		}
		filtered = append(filtered, s)
	}
	srs = filtered

	// Sort if requested
	var sortBy, sortOrder string
	if v := d.Get("sort_by").(string); v != "" {
		sortBy = v
		sortOrder = d.Get("sort_order").(string)
		srs = internal.SortStorageRepositories(srs, sortBy, sortOrder)
	}

	if err := d.Set("srs", srsToMapList(srs)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(internal.Strings([]string{nameLabel, poolId, hostId, srType, sortBy, sortOrder}))
	tflog.Debug(ctx, "Found storage repositories", map[string]interface{}{
		"count": len(srs),
	})

	return nil
}

func srsToMapList(srs []client.StorageRepository) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(srs))
	for _, sr := range srs {
		result = append(result, srToMap(sr))
	}
	return result
}

func resourceSr() *schema.Resource {
	srSchema := resourceSrSchema()
	// The element of a list can only expose computed attributes, so the
	// filterable fields are redeclared as computed here, like name_label is
	// for the pools data source.
	srSchema["name_label"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The name of the storage repository.",
	}
	srSchema["pool_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Id of the pool the storage repository exists on.",
	}
	return &schema.Resource{
		Schema: srSchema,
	}
}
