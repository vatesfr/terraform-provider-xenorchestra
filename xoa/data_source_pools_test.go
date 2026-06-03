package xoa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccXenorchestraDataSource_pools(t *testing.T) {
	resourceName := "data.xenorchestra_pools.pools"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourcePoolsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "pools.#"),
					resource.TestCheckResourceAttr(resourceName, "pools.0.id", accTestPool.Id),
					resource.TestCheckResourceAttr(resourceName, "pools.0.name_label", accTestPool.NameLabel),
					resource.TestCheckResourceAttr(resourceName, "pools.0.master", accTestPool.Master),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_poolsWithTags(t *testing.T) {
	resourceName := "data.xenorchestra_pools.pools_with_tags"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourcePoolsConfigWithTags(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "pools.#"),
					resource.TestCheckResourceAttrSet(resourceName, "pools.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "pools.0.name_label"),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_poolsWithSort(t *testing.T) {
	resourceName := "data.xenorchestra_pools.pools_sorted"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourcePoolsConfigWithSort(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "pools.#"),
				),
			},
		},
	})
}

func testAccXenorchestraDataSourcePoolsConfig() string {
	return `

data "xenorchestra_pools" "pools" {
}
`
}

func testAccXenorchestraDataSourcePoolsConfigWithTags() string {
	return `

data "xenorchestra_pools" "pools_with_tags" {
  tags = ["terraform", "test"]
}
`
}

func testAccXenorchestraDataSourcePoolsConfigWithSort() string {
	return `

data "xenorchestra_pools" "pools_sorted" {
  sort_by = "name_label"
  sort_order = "asc"
}
`
}
