package xoa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraDataSource_pif(t *testing.T) {
	resourceName := "data.xenorchestra_pif.pif"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourcePIFConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourcePIF(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "attached"),
					resource.TestCheckResourceAttr(resourceName, "device", "eth0"),
					resource.TestCheckResourceAttrSet(resourceName, "host"),
					resource.TestCheckResourceAttrSet(resourceName, "network"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttrSet(resourceName, "vlan")),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourcePIF(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find PIF data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("PIF data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourcePIFConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_pif" "pif" {
    device = "eth0"
    vlan = -1
    host_id = "%s"
}
`, accTestPool.Master)
}
