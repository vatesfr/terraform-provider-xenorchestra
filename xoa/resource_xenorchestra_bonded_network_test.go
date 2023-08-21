package xoa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccXOBondedNetwork_createAndPlanWithOmittedPIF(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	resourceName := "xenorchestra_bonded_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraNetworkConfigBonded(nameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttrSet(resourceName, "name_description"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "mtu")),
			},
			{
				// Asserting that pif_ids was set in the previous step failed. This
				// second step is to verify that even though it's not persisted in the
				// state that it still correctly detects changes to the resource when
				// the pif_ids argument is modified.
				Config:             testAccXenorchestraNetworkConfigBondedMissingPIF(nameLabel),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	},
	)
}

func TestAccXOBondedNetwork_updateInPlace(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	resourceName := "xenorchestra_bonded_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	isLocked := "false"
	automatic := "true"
	desc := netDefaultDesc

	updatedNameLabel := nameLabel + " updated"
	updatedDesc := "Non default description"
	updatedAutomatic := "false"
	updatedIsLocked := "true"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraBondedNetworkConfigInPlaceUpdates(nameLabel, desc, automatic, isLocked),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "automatic", automatic),
					resource.TestCheckResourceAttr(resourceName, "default_is_locked", isLocked)),
			},
			{
				Config: testAccXenorchestraBondedNetworkConfigInPlaceUpdates(updatedNameLabel, updatedDesc, updatedAutomatic, updatedIsLocked),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", updatedNameLabel),
					resource.TestCheckResourceAttr(resourceName, "name_description", updatedDesc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "automatic", updatedAutomatic),
					resource.TestCheckResourceAttr(resourceName, "default_is_locked", updatedIsLocked)),
			},
		},
	},
	)
}

func TestAccXOBondedNetwork_updateForceNew(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	resourceName := "xenorchestra_bonded_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	desc := "Non default description"
	origMtu := "950"
	updatedMtu := "1000"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraBondedNetworkConfigNonDefaults(nameLabel, desc, origMtu),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "mtu", origMtu)),
			},
			{
				Config: testAccXenorchestraBondedNetworkConfigNonDefaults(nameLabel, desc, updatedMtu),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "mtu", updatedMtu)),
			},
		},
	},
	)
}

var testAccXenorchestraBondedNetworkPIFs = func() string {
	return fmt.Sprintf(`
data "xenorchestra_pif" "eth1" {
    device = "eth1"
    vlan = -1
    host_id = "%s"
}
data "xenorchestra_pif" "eth2" {
    device = "eth2"
    vlan = -1
    host_id = "%s"
}

`, accTestPIF.Host, accTestPIF.Host)
}

var testAccXenorchestraNetworkConfigBonded = func(name string) string {
	return testAccXenorchestraBondedNetworkPIFs() + fmt.Sprintf(`

resource "xenorchestra_bonded_network" "network" {
    name_label = "%s"
    pool_id = "%s"
    pif_ids = [
      data.xenorchestra_pif.eth1.id,
      data.xenorchestra_pif.eth2.id,
    ]
    bond_mode = "active-backup"
}`, name, accTestPIF.PoolId)
}

var testAccXenorchestraNetworkConfigBondedMissingPIF = func(name string) string {
	return testAccXenorchestraBondedNetworkPIFs() + fmt.Sprintf(`

resource "xenorchestra_bonded_network" "network" {
    name_label = "%s"
    pool_id = "%s"
    pif_ids = [
      data.xenorchestra_pif.eth1.id,
    ]
    bond_mode = "active-backup"
}`, name, accTestPIF.PoolId)
}

var testAccXenorchestraBondedNetworkConfigInPlaceUpdates = func(name, desc, automatic, isLocked string) string {
	return testAccXenorchestraBondedNetworkPIFs() + fmt.Sprintf(`

resource "xenorchestra_bonded_network" "network" {
    name_label = "%s"
    name_description = "%s"
    pool_id = "%s"
    pif_ids = [
      data.xenorchestra_pif.eth1.id,
      data.xenorchestra_pif.eth2.id,
    ]
    bond_mode = "active-backup"
    automatic = %s
    default_is_locked = %s
}`, name, desc, accTestPIF.PoolId, automatic, isLocked)
}

var testAccXenorchestraBondedNetworkConfigNonDefaults = func(name, desc, mtu string) string {
	return testAccXenorchestraBondedNetworkPIFs() + fmt.Sprintf(`
resource "xenorchestra_bonded_network" "network" {
    bond_mode = "active-backup"
    name_label = "%s"
    name_description = "%s"
    pool_id = "%s"
    mtu = %s
    pif_ids = [
      data.xenorchestra_pif.eth1.id,
      data.xenorchestra_pif.eth2.id,
    ]
}
`, name, desc, accTestPIF.PoolId, mtu)
}
