package xoa

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var nonExistantPoolId = "does not exist"

func TestAccXenorchestraDataSource_storageRepository(t *testing.T) {
	resourceName := "data.xenorchestra_sr.sr"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(storageRepositoryConfig{
					nameLabel: accDefaultSr.NameLabel,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceStorageRepository(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "container"),
					resource.TestCheckResourceAttrSet(resourceName, "size"),
					resource.TestCheckResourceAttrSet(resourceName, "physical_usage"),
					resource.TestCheckResourceAttrSet(resourceName, "usage"),
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
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(storageRepositoryConfig{
					nameLabel: "not found",
				}),
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
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(storageRepositoryConfig{
					nameLabel: accDefaultSr.NameLabel,
					poolID:    accDefaultSr.PoolId,
				}),
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
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(storageRepositoryConfig{
					nameLabel: accDefaultSr.NameLabel,
					poolID:    nonExistantPoolId,
				}),
				ExpectError: regexp.MustCompile(`Could not find client.StorageRepository with query`),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_storageRepositoryWithHostId(t *testing.T) {
	resourceName := "data.xenorchestra_sr.sr"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(storageRepositoryConfig{
					poolID: accTestPool.Id,
					hostID: accTestHost.Id,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceStorageRepository(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "container"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accDefaultSr.NameLabel),
					resource.TestCheckResourceAttr(resourceName, "host_id", accTestHost.Id)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_storageRepositoryWithNonExistantHostId(t *testing.T) {
	// Proves the host filter narrows the results: SRs with the default name
	// exist in the test pool, but none live on a host with this id, so the
	// lookup must come back empty instead of returning a match.
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoryConfig(storageRepositoryConfig{
					poolID: accTestPool.Id,
					hostID: "00000000-0000-0000-0000-000000000000",
				}),
				ExpectError: regexp.MustCompile("found `0` srs that match"),
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

type storageRepositoryConfig struct {
	nameLabel string
	poolID    string
	hostID    string
}

func testAccXenorchestraDataSourceStorageRepositoryConfig(config storageRepositoryConfig) string {
	var optionalAttributes strings.Builder
	if config.poolID != "" {
		fmt.Fprintf(&optionalAttributes, "    pool_id = %q\n", config.poolID)
	}
	if config.hostID != "" {
		fmt.Fprintf(&optionalAttributes, "    host_id = %q\n", config.hostID)
	}
	return fmt.Sprintf(`
data "xenorchestra_sr" "sr" {
    name_label = %q
%s    tags = [%q]
}
`, config.nameLabel, optionalAttributes.String(), accTestPrefix)
}
