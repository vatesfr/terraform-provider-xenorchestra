package xoa

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccXenorchestraDataSource_vms(t *testing.T) {
	resourceName := "data.xenorchestra_vms.vms"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceVmsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceVms(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sum")),
			},
		},
	},
	)
}

func testAccXenorchestraDataSourceVmsConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_vms" "vms" {
    pool_id = "%s"
}
`, accTestPool.Id)
}

func testAccCheckXenorchestraDataSourceVms(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Vms data source: %s", n)
		}

		log.Printf("[DEBUG] Found resource again %v", s.RootModule().Resources)
		if rs.Primary.ID == "" {
			return fmt.Errorf("Vms data source ID not set")
		}
		return nil
	}
}
