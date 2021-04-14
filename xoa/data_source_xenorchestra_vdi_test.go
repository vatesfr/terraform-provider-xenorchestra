package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraDataSource_VDI(t *testing.T) {
	resourceName := "data.xenorchestra_vdi.vdi"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceVDIConfig(testIsoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceVDI(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
				),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_vdiNotFound(t *testing.T) {
	resourceName := "data.xenorchestra_vdi.vdi"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceVDIConfig("not found"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceVDI(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
				),
				ExpectError: regexp.MustCompile(`Could not find client.VDI with query`),
			},
		},
	},
	)
}

func testAccXenorchestraDataSourceVDIConfig(nameLabel string) string {
	return fmt.Sprintf(`
data "xenorchestra_vdi" "vdi" {
    name_label = "%s"
    pool_id = "%s"
}
`, nameLabel, accTestPool.Id)
}

func testAccCheckXenorchestraDataSourceVDI(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find VDI data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("VDI data source ID not set")
		}
		return nil
	}
}
