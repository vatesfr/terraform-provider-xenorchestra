package xoa

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccXenorchestraDataSource_hostsSortedDescByNameLabel(t *testing.T) {
	resourceName := "data.xenorchestra_hosts.hosts"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostsConfig("name_label", "desc"),
				Check:  getCompositeAggregateTestFunc(resourceName, "name_label", "desc"),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_hostsSortedAscByNameLabel(t *testing.T) {
	resourceName := "data.xenorchestra_hosts.hosts"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostsConfig("name_label", "asc"),
				Check:  getCompositeAggregateTestFunc(resourceName, "name_label", "asc"),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_hostsSortedDescById(t *testing.T) {
	resourceName := "data.xenorchestra_hosts.hosts"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostsConfig("id", "desc"),
				Check:  getCompositeAggregateTestFunc(resourceName, "id", "desc"),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_hostsSortedAscById(t *testing.T) {
	resourceName := "data.xenorchestra_hosts.hosts"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostsConfig("id", "asc"),
				Check:  getCompositeAggregateTestFunc(resourceName, "id", "asc"),
			},
		},
	},
	)
}

func getCompositeAggregateTestFunc(resourceName, sortBy, sortOrder string) resource.TestCheckFunc {
	attr := fmt.Sprintf("hosts.*.%s", sortBy)
	return resource.ComposeAggregateTestCheckFunc(
		testAccCheckXenorchestraDataSourceHosts(resourceName),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPool.Id),
		// Verify that there are atleast 2 hosts returned
		// This is necessary to test the sorting logic
		resource.TestMatchResourceAttr(resourceName, "hosts.#", regexp.MustCompile(`[2-9]|\d\d\d*`)),
		internal.TestCheckTypeListAttrSorted(resourceName, attr, sortOrder),
		internal.TestCheckTypeSetElemNestedAttrs(resourceName, "hosts.*", map[string]string{
			"pool_id": accTestPool.Id,
		}),
	)
}

func testAccCheckXenorchestraDataSourceHosts(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Hosts data source: %s", n)
		}

		log.Printf("[DEBUG] Found resource again %v", s.RootModule().Resources)
		if rs.Primary.ID == "" {
			return fmt.Errorf("Hosts data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourceHostsConfig(sortBy, sortOrder string) string {
	return fmt.Sprintf(`
data "xenorchestra_hosts" "hosts" {
    pool_id = "%s"
    sort_by = "%s"
    sort_order = "%s"
}
`, accTestPool.Id, sortBy, sortOrder)
}
