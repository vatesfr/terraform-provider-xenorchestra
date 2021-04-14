package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraDataSource_template(t *testing.T) {
	resourceName := "data.xenorchestra_template.template"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceTemplateConfig(accTestPool.Id),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXenorchestraDataSourceTemplate(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
				),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_templateNotFound(t *testing.T) {
	resourceName := "data.xenorchestra_template.template"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceTemplateConfig("not found"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXenorchestraDataSourceTemplate(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
				),
				ExpectError: regexp.MustCompile(`Could not find client.Template with query`),
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

func testAccXenorchestraDataSourceTemplateConfig(poolId string) string {
	return fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
    pool_id = "%s"
}
`, testTemplate.NameLabel, poolId)
}
