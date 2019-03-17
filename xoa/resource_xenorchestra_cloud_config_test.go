package xoa

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccXenorchestraCloudConfig_import(t *testing.T) {
	// resourceName := "xenorchestra_cloud_config.testing"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraCloudConfigDestroy,
		Steps:        []resource.TestStep{},
	})
}

func testAccCheckXenorchestraCloudConfigDestroy(s *terraform.State) error {
	return nil
}
