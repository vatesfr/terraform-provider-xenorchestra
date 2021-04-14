package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var nonExistantPoolId = "does not exist"

func TestAccXenorchestraDataSource_storageRepository(t *testing.T) {
	resourceName := "data.xenorchestra_sr.sr"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(accDefaultSr.NameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceStorageRepository(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accDefaultSr.NameLabel)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_storageRepositoryNotFound(t *testing.T) {
	resourceName := "data.xenorchestra_sr.sr"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig("not found"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceStorageRepository(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accDefaultSr.NameLabel)),
				ExpectError: regexp.MustCompile(`Could not find client.StorageRepository with query`),
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
				Config: testAccXenorchestraDataSourceStorageRepositoryPoolConfig(accDefaultSr.PoolId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceStorageRepository(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accDefaultSr.NameLabel)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_storageRepositoryWithNonExistantPoolId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXenorchestraDataSourceStorageRepositoryPoolConfig(nonExistantPoolId),
				ExpectError: regexp.MustCompile(`Could not find client.StorageRepository with query`),
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

func testAccXenorchestraDataSourceStorageRepositoryConfig(srNameLabel string) string {
	return fmt.Sprintf(`
data "xenorchestra_sr" "sr" {
    name_label = "%s"
    tags = [
	"%s"
    ]
}
`, srNameLabel, accTestPrefix)
}

func testAccXenorchestraDataSourceStorageRepositoryPoolConfig(poolId string) string {
	return fmt.Sprintf(`
data "xenorchestra_sr" "sr" {
    name_label = "%s"
    pool_id = "%s"
    tags = [
	"%s"
    ]
}
`, accDefaultSr.NameLabel, poolId, accTestPrefix)

}
