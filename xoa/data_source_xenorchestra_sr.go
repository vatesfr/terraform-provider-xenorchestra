package xoa

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xo-sdk-go/client"
)

func dataSourceXoaStorageRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStorageRepositoryRead,
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"sr_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"uuid": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"container": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"physical_usage": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"usage": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
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
