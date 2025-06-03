package xoa

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

func TestAccXenorchestraBackup_basic(t *testing.T) {
	resourceName := "xenorchestra_backup.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraBackupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-backup-test"),
					resource.TestCheckResourceAttr(resourceName, "mode", "delta"),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraBackupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-backup-updated"),
					resource.TestCheckResourceAttr(resourceName, "mode", "delta"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckXenorchestraBackupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no backup job ID is set")
		}

		c := testAccProvider.Meta().(*v2.XOClient)
		backup, err := c.Backup().GetJob(context.Background(), rs.Primary.ID, payloads.RestAPIJobQueryVM)
		if err != nil {
			return err
		}

		if backup.ID.String() != rs.Primary.ID {
			return fmt.Errorf("backup job not found")
		}

		return nil
	}
}

func testAccCheckXenorchestraBackupDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*v2.XOClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_backup" {
			continue
		}

		_, err := c.Backup().GetJob(context.Background(), rs.Primary.ID, payloads.RestAPIJobQueryVM)
		if err == nil {
			return fmt.Errorf("backup job still exists")
		}

		if err.Error() != fmt.Sprintf("backup job not found with id: %s", rs.Primary.ID) {
			return fmt.Errorf("expected 'backup job not found' error, got %s", err)
		}
	}

	return nil
}

func testAccBackupConfig() string {
	// For testing, we need to reference an existing VM.
	// In a real environment, users would specify their own VM IDs.
	// For now, just use a placeholder that will be replaced in actual testing.
	return `
resource "xenorchestra_backup" "test" {
  name = "terraform-backup-test"
  mode = "delta"
  vms  = ["00000000-0000-0000-0000-000000000000"]  # Will be replaced in actual test
  
  schedule {
    cron     = "0 0 * * *"  # Daily at midnight
    enabled  = true
    name     = "terraform-test-schedule"
    timezone = "UTC"
  }
  
  settings {
    compression_enabled   = true
    offline_backup        = false
    checkpoint_snapshot   = false
    remote_enabled        = false
    remote_retention      = 0
    report_when_fail_only = true
    
    long_term_retention {
      daily {
        retention = 2
      }
      weekly {
        retention = 2
      }
      monthly {
        retention = 2
      }
      yearly {
        retention = 1
      }
    }
  }
}
`
}

func testAccBackupConfigUpdated() string {
	return `
resource "xenorchestra_backup" "test" {
  name = "terraform-backup-updated"
  mode = "delta"
  vms  = ["00000000-0000-0000-0000-000000000000"]  # Will be replaced in actual test
  
  schedule {
    cron     = "0 0 * * *"
    enabled  = false
    name     = "terraform-test-schedule-updated"
    timezone = "UTC"
  }
  
  settings {
    compression_enabled   = true
    offline_backup        = true
    checkpoint_snapshot   = true
    remote_enabled        = false
    remote_retention      = 0
    report_when_fail_only = true
    
    long_term_retention {
      daily {
        retention = 2
      }
      weekly {
        retention = 2
      }
      monthly {
        retention = 2
      }
      yearly {
        retention = 1
      }
    }
  }
}
`
}
