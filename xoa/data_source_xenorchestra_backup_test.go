package xoa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccXenorchestraDataSource_backup(t *testing.T) {
	resourceName := "data.xenorchestra_backup.backup"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceBackupConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceBackup(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

func testAccXenorchestraDataSourceBackupConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_backup" "backup" {
    name = "%s"
}
`, testAccBackupName)
}

var testAccBackupName = "test-backup-job"

func testAccCheckXenorchestraDataSourceBackup(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find Backup data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("backup data source ID not set")
		}
		return nil
	}
}
