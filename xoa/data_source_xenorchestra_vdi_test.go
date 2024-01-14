package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccXenorchestraDataSource_vdiById(t *testing.T) {
	resourceName := "data.xenorchestra_vdi.vdi"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceVDIConfigById(testIso.VDIId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceVDI(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "parent", ""),
				),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_vdiByNameLabel(t *testing.T) {
	resourceName := "data.xenorchestra_vdi.vdi"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceVDIConfig(testIso.NameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceVDI(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "parent", ""),
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

func TestAccXenorchestraDataSource_vdiExactlyOneOfIdOrNameLabelRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
data "xenorchestra_vdi" "vdi" {
    id = "test"
    name_label = "test"
}
			    `,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	},
	)
}

func testAccXenorchestraDataSourceVDIConfigById(id string) string {
	return fmt.Sprintf(`
data "xenorchestra_vdi" "vdi" {
    id = "%s"
}
`, id)
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
