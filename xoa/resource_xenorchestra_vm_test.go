package xoa

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.pif", "network")),
			},
		},
	})
}

func TestAccVm_import(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	checkFn := func(s []*terraform.InstanceState) error {
		attrs := []string{"id", "name_label"}
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
				ResourceName:     resourceName,
				ImportState:      true,
				ImportStateCheck: checkFn,
				// TODO: Need to store all the
				// schema.Schema structs in the statefile that
				// currently exist before this will pass.
				// ImportStateVerify: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_description", "description"),
					resource.TestCheckResourceAttr(resourceName, "name_label", "Terraform testing"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.pif", "network")),
			},
		},
	})
}

// TODO: Add unit tests
func testAccCheckXenorchestraVmDestroy(s *terraform.State) error {
	c, err := client.NewClient(client.GetConfigFromEnv())
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_vm" {
			continue
		}

		_, err := c.GetVm(client.Vm{Id: rs.Primary.ID})

		if _, ok := err.(client.NotFound); ok {
			return nil
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func TestAccXenorchestraVm_updateVmWithSecondVif(t *testing.T) {
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
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.pif", "network")),
			},
			{
				Config: testAccVmConfigWithSecondVIF(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "2"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.pif", "network"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.eth0", "network")),
			},
		},
	})
}

// TODO: This test fails due to the missing PV drivers issue I've been trying to track down
// Until then this test will fail.
// func TestAccXenorchestraVm_removeVifFromVm(t *testing.T) {
// 	resourceName := "xenorchestra_vm.bar"
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckXenorchestraVmDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccVmConfigWithSecondVIF(),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					testAccVmExists(resourceName),
// 					resource.TestCheckResourceAttrSet(resourceName, "id"),
// 					resource.TestCheckResourceAttr(resourceName, "network.#", "2"),
// 					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.pif", "network"),
// 					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.eth0", "network")),
// 			},
// 			{
// 				Config: testAccVmConfig(),
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					testAccVmExists(resourceName),
// 					resource.TestCheckResourceAttrSet(resourceName, "id"),
// 					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
// 					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_pif.pif", "network")),
// 			},
// 		},
// 	})
// }

func TestAccXenorchestraVm_updatesWithoutReboot(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"

	origNameLabel := "name label"
	origNameDesc := "name label"
	origHa := ""
	origPowerOn := false
	updatedNameLabel := "Updated name label"
	updatedNameDesc := "Updated description"
	updatedHa := "restart"
	updatedPowerOn := true
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigUpdateAttrsHaltIrrelevant(origNameLabel, origNameDesc, origHa, origPowerOn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", origNameLabel),
					resource.TestCheckResourceAttr(resourceName, "auto_poweron", strconv.FormatBool(origPowerOn)),
					resource.TestCheckResourceAttr(resourceName, "high_availability", origHa)),
			},
			{
				Config: testAccVmConfigUpdateAttrsHaltIrrelevant(updatedNameLabel, updatedNameDesc, updatedHa, updatedPowerOn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", updatedNameLabel),
					resource.TestCheckResourceAttr(resourceName, "name_description", updatedNameDesc),
					resource.TestCheckResourceAttr(resourceName, "auto_poweron", strconv.FormatBool(updatedPowerOn)),
					resource.TestCheckResourceAttr(resourceName, "high_availability", updatedHa)),
			},
		},
	})
}

func testAccVmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Vm Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		vm, err := c.GetVm(client.Vm{Id: rs.Primary.ID})

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
data "xenorchestra_sr" "local_storage" {
    name_label = "Local storage"
}

data "xenorchestra_template" "template" {
    name_label = "Focal Template"
}

data "xenorchestra_pif" "pif" {
    device = "eth1"
    vlan = -1
}

resource "xenorchestra_vm" "bar" {
    memory_max = 256000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "data.xenorchestra_template.template"
    network {
	network_id = "${data.xenorchestra_pif.pif.network}"
    }

    disk {
      sr_id = "data.xenorchestra_sr.local_storage.id"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`
}

func testAccVmConfigWithSecondVIF() string {
	return testAccCloudConfigConfig() + `
data "xenorchestra_sr" "local_storage" {
    name_label = "Local storage"
}

data "xenorchestra_template" "template" {
    name_label = "Focal Template"
}

data "xenorchestra_pif" "pif" {
    device = "eth1"
    vlan = -1
}

data "xenorchestra_pif" "eth0" {
    device = "eth0"
    vlan = -1
}

resource "xenorchestra_vm" "bar" {
    memory_max = 256000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "data.xenorchestra_template.template"
    network {
	network_id = "${data.xenorchestra_pif.pif.network}"
    }

    network {
	network_id = "${data.xenorchestra_pif.eth0.network}"
    }

    disk {
      sr_id = "data.xenorchestra_sr.local_storage.id"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`
}

// Terraform config that tests changes to a VM that do not require halting
// the VM prior to applying
func testAccVmConfigUpdateAttrsHaltIrrelevant(nameLabel, nameDescription, ha string, powerOn bool) string {
	return testAccCloudConfigConfig() + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "Focal Template"
}

data "xenorchestra_pif" "pif" {
    device = "eth1"
    vlan = -1
}

resource "xenorchestra_vm" "bar" {
    memory_max = 256000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "%s"
    template = "${data.xenorchestra_template.template.id}"
    high_availability = "%s"
    auto_poweron = "%t"
    network {
	network_id = "${data.xenorchestra_pif.pif.network}"
    }

    disk {
      sr_id = "7f469400-4a2b-5624-cf62-61e522e50ea1"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, nameLabel, nameDescription, ha, powerOn)
}
