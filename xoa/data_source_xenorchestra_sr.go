package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXoaStorageRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStorageRepositoryRead,
		Description: `Provides information about a Storage repository to ease the lookup of VM storage information.

**Note:** If there are multiple storage repositories that match terraform will fail.
Ensure that your name_label, pool_id and tags identify a unique storage repository.`,
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
				Description: "The storage size.",
				Computed:    true,
			},
			"physical_usage": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "The physical storage size.",
				Computed:    true,
			},
			"usage": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "The current usage for this storage repository.",
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

	l := len(srs)
	if l != 1 {
		return errors.New(fmt.Sprintf("found `%d` srs that match %+v. Storage repositories must be uniquely named to use this data source", l, srs))
	}

	sr = srs[0]

	d.SetId(sr.Id)
	d.Set("sr_type", sr.SRType)
	d.Set("uuid", sr.Uuid)
	d.Set("pool_id", sr.PoolId)
	d.Set("size", sr.Size)
	d.Set("physical_usage", sr.PhysicalUsage)
	d.Set("usage", sr.Usage)
	d.Set("container", sr.Container)
	return nil
}

func tagsFromInterfaceSlice(values []interface{}) []string {
	s := make([]string, 0, len(values))

	for _, value := range values {
		s = append(s, value.(string))
	}
	return s
}
