package xoa

import (
	"fmt"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraCloudConfig_readAfterDelete(t *testing.T) {
	resourceName := "xenorchestra_cloud_config.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraCloudConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudConfigConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudConfigExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
			{
				Config:             testAccCloudConfigConfig(),
				Check:              testAccCheckXenorchestraCloudConfigDestroyNow(resourceName),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccCloudConfigConfig(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccXenorchestraCloudConfig_create(t *testing.T) {
	resourceName := "xenorchestra_cloud_config.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraCloudConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudConfigConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudConfigExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
		},
	})
}

func TestAccXenorchestraCloudConfig_updateName(t *testing.T) {
	resourceName := "xenorchestra_cloud_config.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraCloudConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudConfigConfig(),
			},
			{
				Config: testAccCloudConfigConfigUpdatedName(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "updated")),
			},
		},
	})
}

func TestAccXenorchestraCloudConfig_updateTemplate(t *testing.T) {
	resourceName := "xenorchestra_cloud_config.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraCloudConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudConfigConfig(),
			},
			{
				Config: testAccCloudConfigConfigUpdatedTemplate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "template", "updated")),
			},
		},
	})
}

func TestAccXenorchestraCloudConfig_import(t *testing.T) {
	resourceName := "xenorchestra_cloud_config.bar"
	// TODO: Need to figure out how to get this to make sure all the attrs
	// are set. Right now it doesn't actually provide much protection
	checkFn := func(s []*terraform.InstanceState) error {
		attrs := []string{"id", "name", "template"}
		for _, attr := range attrs {
			_, ok := s[0].Attributes[attr]

			if !ok {
				return fmt.Errorf("attribute %s should be set", attr)
			}
		}
		return nil
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraCloudConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudConfigConfig(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudConfigExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
		},
	})
}

func testAccCloudConfigConfig() string {
	return `
resource "xenorchestra_cloud_config" "bar" {
    name = "testing"
    template = "template"
}
`
}

func testAccCloudConfigConfigUpdatedName() string {
	return `
resource "xenorchestra_cloud_config" "bar" {
    name = "updated"
    template = "template"
}
`
}

func testAccCloudConfigConfigUpdatedTemplate() string {
	return `
resource "xenorchestra_cloud_config" "bar" {
    name = "name"
    template = "updated"
}
`
}

func testAccCloudConfigExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudConfig Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		config, err := c.GetCloudConfig(rs.Primary.ID)

		if config.Id == rs.Primary.ID {
			return nil
		}
		return nil
	}
}

func testAccCheckXenorchestraCloudConfigDestroyNow(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudConfig Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		err = c.DeleteCloudConfig(rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckXenorchestraCloudConfigDestroy(s *terraform.State) error {
	c, err := client.NewClient(client.GetConfigFromEnv())
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_cloud_config" {
			continue
		}

		config, err := c.GetCloudConfig(rs.Primary.ID)

		if err != nil {
			return err
		}

		if config != nil {
			return fmt.Errorf("Cloud config (%s) still exists", config.Id)
		}
	}
	return nil
}
