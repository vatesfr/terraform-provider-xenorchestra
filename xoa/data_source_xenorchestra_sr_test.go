package xoa

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var nameLabel = "XenServer Tools"
var poolId = "cadf25ab-91ff-6fc0-041f-5a7033c4bc78"
var nonExistantPoolId = "does not exist"

func TestAccXenorchestraDataSource_storageRepository(t *testing.T) {
	resourceName := "data.xenorchestra_sr.sr"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceStorageRepository(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "name_label", nameLabel)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_storageRepositoryWithPoolId(t *testing.T) {
	resourceName := "data.xenorchestra_sr.sr"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryPoolConfig(poolId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceStorageRepository(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "name_label", nameLabel)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_storageRepositoryWithNonExistantPoolId(t *testing.T) {
	resourceName := "data.xenorchestra_sr.sr"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryPoolConfig(nonExistantPoolId),
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return fmt.Errorf("Can't find StorageRepository data source: %s", resourceName)
					}

					if rs.Primary.ID == "" {
						return nil
					}

					return errors.New("Storage Repository data source should have failed to be found")
				},
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceStorageRepository(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find StorageRepository data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("StorageRepository data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourceStorageRepositoryConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_sr" "sr" {
    name_label = "%s"
}
`, nameLabel)
}

func testAccXenorchestraDataSourceStorageRepositoryPoolConfig(poolId string) string {
	return fmt.Sprintf(`
data "xenorchestra_sr" "sr" {
    name_label = "%s"
    pool_id = "%s"
}
`, nameLabel, poolId)

}
