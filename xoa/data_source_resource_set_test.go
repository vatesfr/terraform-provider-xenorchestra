package xoa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var resourceSetName = "xenserver-ddelnano"

func TestAccXenorchestraDataSource_resourceSet(t *testing.T) {
	resourceName := "data.xenorchestra_resource_set.rs"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceResourceSetConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceResourceSet(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-resource-set")),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceResourceSet(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find ResourceSet data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ResourceSet data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourceResourceSetConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_resource_set" "rs" {
    name = "%s"
}
`, testResourceSetName)
}
