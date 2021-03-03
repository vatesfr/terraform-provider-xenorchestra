package xoa

import (
	"fmt"
	"log"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraDataSource_hosts(t *testing.T) {
	resourceName := "data.xenorchestra_hosts.hosts"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceHosts(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPool.Id),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "hosts.*", map[string]string{
						"pool_id": accTestPool.Id,
					}),
				),
			},
		},
	},
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

func testAccXenorchestraDataSourceHostsConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_hosts" "hosts" {
    pool_id = "%s"
}
`, accTestPool.Id)
}
