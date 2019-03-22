package xoa

import (
	"fmt"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccXenorchestraVm_create(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
		},
	})
}

func testAccVm_import(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
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
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_description", "description"),
					resource.TestCheckResourceAttr(resourceName, "name_label", "Name")),
			},
		},
	})
}

// TODO: Add unit tests
func testAccCheckXenorchestraVmDestroy(s *terraform.State) error {
	c, err := client.NewClient()
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_vm" {
			continue
		}

		_, err := c.GetVm(rs.Primary.ID)

		if _, ok := err.(client.NotFound); ok {
			return nil
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: Add vm update tests
// test case updating cloud config recreates vm

// TODO: Add unit tests
func testAccVmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Vm Id is set")
		}

		c, err := client.NewClient()
		if err != nil {
			return err
		}

		vm, err := c.GetVm(rs.Primary.ID)

		if err != nil {
			return err
		}

		if vm.Id == rs.Primary.ID {
			return nil
		}
		return nil
	}
}

func testAccVmConfig() string {
	return testAccCloudConfigConfig() + `
data "xenorchestra_template" "template" {
    name_label = "Puppet agent - Bionic 18.04 - 1"
}

data "xenorchestra_pif" "pif" {
    device = "eth1"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 1073733632
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Name"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_pif.pif.network}"
    }

    disk {
      sr_id = "7f469400-4a2b-5624-cf62-61e522e50ea1"
      name_label = "Ubuntu Bionic Beaver 18.04_imavo"
      size = 32212254720 
    }
}
`
}
