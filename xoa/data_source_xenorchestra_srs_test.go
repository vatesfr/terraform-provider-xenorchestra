package xoa

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccXenorchestraDataSource_storageRepositories(t *testing.T) {
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "srs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.id"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.name_label"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.sr_type"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.container"),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_storageRepositoriesWithPoolId(t *testing.T) {
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{
					poolID: accTestPool.Id,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "srs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "srs.0.pool_id", accTestPool.Id),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_storageRepositoriesWithHostId(t *testing.T) {
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{
					poolID: accTestPool.Id,
					hostID: accTestHost.Id,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "srs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "srs.0.container", accTestHost.Id),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_storageRepositoriesWithSrType(t *testing.T) {
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{
					srType: accIsoSr.SRType,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "srs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "srs.0.sr_type", accIsoSr.SRType),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_storageRepositoriesWithTags(t *testing.T) {
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{
					tags: []string{accTestPrefix},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "srs.#"),
					resource.TestCheckResourceAttrSet(resourceName, "srs.0.id"),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_storageRepositoriesWithSort(t *testing.T) {
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{
					sortBy:    "size",
					sortOrder: "desc",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "srs.#"),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_storageRepositoriesNotFound(t *testing.T) {
	// Proves the filters narrow the results: a name that does not exist
	// yields an empty list instead of any match.
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{
					nameLabel: "not found",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srs.#", "0"),
				),
			},
		},
	})
}

func TestAccXenorchestraDataSource_storageRepositoriesWithNonExistantHostId(t *testing.T) {
	// Proves the host filter narrows the results: SRs exist in the test
	// pool, but none live on a host with this id.
	resourceName := "data.xenorchestra_srs.srs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceStorageRepositoriesConfig(srsConfig{
					poolID: accTestPool.Id,
					hostID: "00000000-0000-0000-0000-000000000000",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "srs.#", "0"),
				),
			},
		},
	})
}

type srsConfig struct {
	nameLabel string
	poolID    string
	hostID    string
	srType    string
	tags      []string
	sortBy    string
	sortOrder string
}

func testAccXenorchestraDataSourceStorageRepositoriesConfig(config srsConfig) string {
	var optionalAttributes strings.Builder
	if config.nameLabel != "" {
		fmt.Fprintf(&optionalAttributes, "    name_label = %q\n", config.nameLabel)
	}
	if config.poolID != "" {
		fmt.Fprintf(&optionalAttributes, "    pool_id = %q\n", config.poolID)
	}
	if config.hostID != "" {
		fmt.Fprintf(&optionalAttributes, "    host_id = %q\n", config.hostID)
	}
	if config.srType != "" {
		fmt.Fprintf(&optionalAttributes, "    sr_type = %q\n", config.srType)
	}
	if len(config.tags) > 0 {
		quoted := make([]string, len(config.tags))
		for i, t := range config.tags {
			quoted[i] = fmt.Sprintf("%q", t)
		}
		fmt.Fprintf(&optionalAttributes, "    tags = [%s]\n", strings.Join(quoted, ", "))
	}
	if config.sortBy != "" {
		fmt.Fprintf(&optionalAttributes, "    sort_by = %q\n", config.sortBy)
	}
	if config.sortOrder != "" {
		fmt.Fprintf(&optionalAttributes, "    sort_order = %q\n", config.sortOrder)
	}
	return fmt.Sprintf(`
data "xenorchestra_srs" "srs" {
%s}
`, optionalAttributes.String())
}
