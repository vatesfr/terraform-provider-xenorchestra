package xoa

import (
	"errors"
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
		},
	}
}

func dataSourceStorageRepositoryRead(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)

	if err != nil {
		return err
	}

	nameLabel := d.Get("name_label").(string)
	poolId, found := d.GetOk("pool_id")

	sr := client.StorageRepository{
		NameLabel: nameLabel,
	}

	if found {
		sr.PoolId = poolId.(string)
	}

	srs, err := c.GetStorageRepository(sr)

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

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
	return nil
}
