package xoa

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraVm_createAndPlanWithNonExistantVm(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	removeVm := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		vm, err := c.GetVm(client.Vm{
			NameLabel: "Terraform testing",
		})

		if err != nil {
			t.Fatalf("failed to find VM with error: %v", err)
		}

		err = c.DeleteVm(vm.Id)
		if err != nil {
			t.Fatalf("failed to delete VM with error: %v", err)
		}
	}
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
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
				Destroy: false,
			},
			{
				PreConfig:          removeVm,
				Config:             testAccVmConfig(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccXenorchestraVm_createWithoutCloudConfig(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmWithoutCloudInitConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithMacAddress(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	macAddress := "00:0a:83:b1:c0:83"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithMacAddress(macAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "network.*", map[string]string{
						"mac_address": macAddress,
					}),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_disconnectAttachedVif(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmVifAttachedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config:             testAccVmVifDetachedConfig(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVmVifDetachedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "false"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_attachDisconnectedVif(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	removeVif := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		vm, err := c.GetVm(client.Vm{
			NameLabel: "Terraform testing",
		})

		if err != nil {
			t.Fatalf("failed to find VM with error: %v", err)
		}
		time.Sleep(30 * time.Second)

		err = c.DisconnectVIF(&client.VIF{Id: vm.VIFs[0]})
		if err != nil {
			t.Fatalf("failed to disconnect VIF with error: %v", err)
		}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmVifAttachedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				PreConfig: removeVif,
				Config:    testAccVmVifAttachedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "network.*", map[string]string{
						"attached": "false",
						"device":   "0",
					}),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVmVifAttachedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "network.*", map[string]string{
						"attached": "true",
						"device":   "0",
					}),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_import(t *testing.T) {
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
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
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
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config: testAccVmConfigWithSecondVIF(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "2"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.0.*", "data.xenorchestra_network.network", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.device", "0"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "true"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.1.*", "data.xenorchestra_network.network2", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.1.device", "1"),
					resource.TestCheckResourceAttr(resourceName, "network.1.attached", "true")),
			},
		},
	})
}

func TestAccXenorchestraVm_removeVifFromVmAndDeviceOrdering(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithThreeVIFs(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "3"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.0.*", "data.xenorchestra_network.network", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.device", "0"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "true"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.1.*", "data.xenorchestra_network.network2", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.1.device", "1"),
					resource.TestCheckResourceAttr(resourceName, "network.1.attached", "true"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.2.*", "data.xenorchestra_network.network2", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.2.device", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.2.attached", "true")),
			},
			{
				Config: testAccVmConfigWithSecondVIF(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "2"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.device", "0"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "true"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.1.*", "data.xenorchestra_network.network", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.1.device", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.1.attached", "true")),
			},
		},
	})
}

func TestAccXenorchestraVm_updatesWithoutReboot(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"

	origNameLabel := "name label"
	origNameDesc := "name label"
	origHa := ""
	origPowerOn := false
	updatedNameLabel := "Terraform Updated name label"
	updatedNameDesc := "Terraform Updated description"
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

func TestAccXenorchestraVm_createAndUpdateWithResourceSet(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithResourceSet(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_set"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config: testAccVmConfigWithoutResourceSet(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "resource_set", ""),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
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

func testAccVmWithoutCloudInitConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
    pool_id = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%[2]s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    name_label = "Terraform testing"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, accDefaultSr.Id)
}

func testAccVmConfig() string {
	return testAccCloudConfigConfig("vm-template", "template") + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, accDefaultSr.Id)
}

func testAccVmVifAttachedConfig() string {
	return testAccCloudConfigConfig("vm-template", "template") + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
	attached = true
	device = "0"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, accDefaultSr.Id)
}

func testAccVmVifDetachedConfig() string {
	return testAccCloudConfigConfig("vm-template", "template") + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
	attached = false
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, accDefaultSr.Id)
}

func testAccVmConfigWithMacAddress(macAddress string) string {
	return testAccCloudConfigConfig("vm-template", "template") + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
	mac_address = "%s"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, macAddress, accDefaultSr.Id)
}

func testAccVmConfigWithSecondVIF() string {
	return testAccCloudConfigConfig("vm-template", "template") + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

data "xenorchestra_network" "network2" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth1"
    pool_id = "%[2]s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }
    network {
	network_id = "${data.xenorchestra_network.network2.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, accDefaultSr.Id)
}

func testAccVmConfigWithThreeVIFs() string {
	return testAccCloudConfigConfig("vm-template", "template") + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

data "xenorchestra_network" "network2" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth1"
    pool_id = "%[2]s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }
    network {
	network_id = "${data.xenorchestra_network.network2.id}"
    }
    network {
	network_id = "${data.xenorchestra_network.network2.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, accDefaultSr.Id)
}

// Terraform config that tests changes to a VM that do not require halting
// the VM prior to applying
func testAccVmConfigUpdateAttrsHaltIrrelevant(nameLabel, nameDescription, ha string, powerOn bool) string {
	return testAccCloudConfigConfig("vm-template", "template") + fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "%s"
    template = "${data.xenorchestra_template.template.id}"
    high_availability = "%s"
    auto_poweron = "%t"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accTemplateName, accTestPool.Id, nameLabel, nameDescription, ha, powerOn, accDefaultSr.Id)
}

func testAccVmConfigWithResourceSet() string {
	return testAccCloudConfigConfig("vm-template", "template") + testAccVmResourceSet() + fmt.Sprintf(`

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing resource sets"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    resource_set = "${xenorchestra_resource_set.rs.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accDefaultSr.Id)
}

func testAccVmResourceSet() string {
	return fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    // TODO: Replace this with a better solution
    name_label = "Pool-wide network associated with eth0"
    pool_id = "%s"
}

resource "xenorchestra_resource_set" "rs" {
    name = "terraform-vm-acceptance-test"
    subjects = []
    objects = [
	"${data.xenorchestra_template.template.id}",
	"%s",
	"${data.xenorchestra_network.network.id}",
    ]

    limit {
      type = "cpus"
      quantity = 20
    }

    limit {
      type = "disk"
      quantity = 107374182400
    }

    limit {
      type = "memory"
      quantity = 12884901888
    }
}
`, accTemplateName, accTestPool.Id, accDefaultSr.Id)
}

func testAccVmConfigWithoutResourceSet() string {
	return testAccCloudConfigConfig("vm-template", "template") + testAccVmResourceSet() + fmt.Sprintf(`

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Terraform testing resource sets"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "xo provider root"
      size = 10000000000
    }
}
`, accDefaultSr.Id)
}
