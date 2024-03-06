package xoa

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/vatesfr/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccXenorchestraDataSource_cloudConfig(t *testing.T) {
	resourceName := "data.xenorchestra_cloud_config.config"
	cloudConfigName := fmt.Sprintf("%s-cloud-config-data-source-test - %s", accTestPrefix, t.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: createCloudConfig(t, cloudConfigName),
				Config:    testAccXenorchestraDataSourceCloudConfigConfig(cloudConfigName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceCloudConfig(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_cloudConfigProvidesErrorWhenNotFound(t *testing.T) {
	cloudConfigName := "Does not exist"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXenorchestraDataSourceCloudConfigConfig(cloudConfigName),
				ExpectError: regexp.MustCompile("Could not find client.CloudConfig with query"),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_cloudConfigErrorsWhenNonUniqueNamesExist(t *testing.T) {
	cloudConfigName := resource.PrefixedUniqueId(fmt.Sprintf("%s-cloud-config-data-source-test", accTestPrefix))
	c := createCloudConfig(t, cloudConfigName)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					for i := 0; i < 2; i++ {
						c()
					}
				},
				Config:      testAccXenorchestraDataSourceCloudConfigConfig(cloudConfigName),
				ExpectError: regexp.MustCompile("found `2` cloud configs with name"),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceCloudConfig(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find CloudConfig data source: %s", n)
		}

		log.Printf("[DEBUG] Found resource again %v", s.RootModule().Resources)
		if rs.Primary.ID == "" {
			return fmt.Errorf("CloudConfig data source ID not set")
		}
		return nil
	}
}

func createCloudConfig(t *testing.T, name string) func() {
	return func() {
		c, err := client.NewClient(client.GetConfigFromEnv())

		if err != nil {
			t.Fatalf("failed to created client with error: %v", err)
		}

		_, err = c.CreateCloudConfig(name, "Terraform acceptance testing template")

		if err != nil {
			t.Fatalf("failed to create cloud config for test with error: %v", err)
		}
	}
}

func testAccXenorchestraDataSourceCloudConfigConfig(cloudConfigName string) string {
	return fmt.Sprintf(`
data "xenorchestra_cloud_config" "config" {
    name = "%s"
}
`, cloudConfigName)
}
