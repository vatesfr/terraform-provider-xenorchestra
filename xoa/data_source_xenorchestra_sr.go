package xoa

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaStorageRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStorageRepositoryRead,
		Description: `Provides information about a Storage repository to ease the lookup of VM storage information.

**Note:** If there are multiple storage repositories that match terraform will fail.
Ensure that your name_label, pool_id, host_id and tags identify a unique storage repository.`,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Description: "The name of the storage repository to look up",
				Type:        schema.TypeString,
				Required:    true,
			},
			"sr_type": &schema.Schema{
				Description: "The type of storage repository (lvm, udev, iso, user, etc).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pool_id": &schema.Schema{
				Description: "The Id of the pool the storage repository exists on.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"host_id": &schema.Schema{
				Description: "The Id of the host the storage repository exists on. For host-local storage repositories the SR's `container` is the host itself, so this filters the repositories down to a single host when several hosts in the pool share the same SR name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"uuid": &schema.Schema{
				Type:        schema.TypeString,
				Description: "uuid of the storage repository. This is equivalent to the id.",
				Computed:    true,
			},
			"container": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The storage container.",
				Computed:    true,
			},
			"size": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "The total storage size in bytes.",
				Computed:    true,
			},
			"physical_usage": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "The physical storage usage in bytes.",
				Computed:    true,
			},
			"usage": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "The current storage usage in bytes.",
				Computed:    true,
			},
			"tags": resourceTags(),
		},
	}
}

func dataSourceStorageRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	nameLabel := d.Get("name_label").(string)
	poolId := d.Get("pool_id").(string)
	hostId := d.Get("host_id").(string)
	tags := d.Get("tags").(*schema.Set).List()

	sr := client.StorageRepository{
		NameLabel: nameLabel,
		PoolId:    poolId,
		Tags:      tagsFromInterfaceSlice(tags),
	}

	srs, err := c.GetStorageRepository(sr)

	if err != nil {
		return err
	}

	// The SDK cannot filter by host, but for host-local SRs the SR's
	// `container` is the host that owns it, so filter by container here.
	if hostId != "" {
		filtered := make([]client.StorageRepository, 0, len(srs))
		for _, s := range srs {
			if s.Container == hostId {
				filtered = append(filtered, s)
			}
		}
		srs = filtered
	}

	l := len(srs)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` srs that match %+v. Storage repositories must be uniquely named to use this data source", l, srs))
	}

	sr = srs[0]

	d.SetId(sr.Id)
	for k, v := range srToMap(sr) {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func srToMap(sr client.StorageRepository) map[string]interface{} {
	return map[string]interface{}{
		"id":             sr.Id,
		"uuid":           sr.Uuid,
		"name_label":     sr.NameLabel,
		"pool_id":        sr.PoolId,
		"sr_type":        sr.SRType,
		"container":      sr.Container,
		"size":           sr.Size,
		"physical_usage": sr.PhysicalUsage,
		"usage":          sr.Usage,
		"tags":           sr.Tags,
	}
}

func tagsFromInterfaceSlice(values []interface{}) []string {
	s := make([]string, 0, len(values))

	for _, value := range values {
		s = append(s, value.(string))
	}
	return s
}
