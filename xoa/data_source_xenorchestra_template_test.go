package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccXenorchestraDataSource_template(t *testing.T) {
	resourceName := "data.xenorchestra_template.template"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceTemplateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXenorchestraDataSourceTemplate(resourceName),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile("^OpaqueRef:")),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
				),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceTemplate(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Template data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Template data source ID not set")
		}
		return nil
	}
}

const testAccXenorchestraDataSourceTemplateConfig = `
data "xenorchestra_template" "template" {
    name_label = "Asianux Server 4 (64-bit)"
}
`
