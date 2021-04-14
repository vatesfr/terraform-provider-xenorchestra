package xoa

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraDataSource_pool(t *testing.T) {
	resourceName := "data.xenorchestra_pool.pool"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourcePoolConfig(accTestPool.NameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourcePool(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "master"),
					resource.TestCheckResourceAttr(resourceName, "cpus.%", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "cpus.sockets"),
					resource.TestCheckResourceAttrSet(resourceName, "cpus.cores"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accTestPool.NameLabel)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_poolNotFound(t *testing.T) {
	resourceName := "data.xenorchestra_pool.pool"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourcePoolConfig("not found"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourcePool(resourceName),
				),
				ExpectError: regexp.MustCompile(`Could not find client.Pool with query`),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourcePool(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Pool data source: %s", n)
		}

		log.Printf("[DEBUG] Found resource again %v", s.RootModule().Resources)
		if rs.Primary.ID == "" {
			return fmt.Errorf("Pool data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourcePoolConfig(pool string) string {
	return fmt.Sprintf(`
data "xenorchestra_pool" "pool" {
    name_label = "%s"
}
`, pool)
}
