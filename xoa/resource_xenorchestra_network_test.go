package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccXONetwork_create(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	resourceName := "xenorchestra_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraNetworkConfig(nameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttrSet(resourceName, "name_description"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "mtu"),
					resource.TestCheckResourceAttrSet(resourceName, "vlan")),
			},
		},
	},
	)
}

func TestAccXONetwork_createWithVlanRequiresPIF(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccXenorchestraNetworkConfigWithoutVlan(nameLabel),
				ExpectError: regexp.MustCompile("all of `pif_id,vlan` must be specified"),
			},
			{
				Config:      testAccXenorchestraNetworkConfigWithoutPIF(nameLabel),
				ExpectError: regexp.MustCompile("all of `pif_id,vlan` must be specified"),
			},
		},
	},
	)
}

func TestAccXONetwork_createWithNonDefaults(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	resourceName := "xenorchestra_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	desc := "Non default description"
	nbd := "true"
	mtu := "950"
	vlan := "22"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraNetworkConfigNonDefaultsWithVlan(nameLabel, desc, mtu, nbd, vlan),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "mtu", mtu),
					resource.TestCheckResourceAttr(resourceName, "vlan", vlan),
					resource.TestCheckResourceAttr(resourceName, "nbd", nbd)),
			},
		},
	},
	)
}

func TestAccXONetwork_updateInPlace(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	resourceName := "xenorchestra_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	isLocked := "false"
	automatic := "true"
	desc := netDefaultDesc
	nbd := "false"

	updatedNameLabel := nameLabel + " updated"
	updatedDesc := "Non default description"
	updatedNbd := "true"
	updatedAutomatic := "false"
	updatedIsLocked := "true"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraNetworkConfigInPlaceUpdates(nameLabel, desc, nbd, automatic, isLocked),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "automatic", automatic),
					resource.TestCheckResourceAttr(resourceName, "default_is_locked", isLocked),
					resource.TestCheckResourceAttr(resourceName, "nbd", nbd)),
			},
			{
				Config: testAccXenorchestraNetworkConfigInPlaceUpdates(updatedNameLabel, updatedDesc, updatedNbd, updatedAutomatic, updatedIsLocked),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", updatedNameLabel),
					resource.TestCheckResourceAttr(resourceName, "name_description", updatedDesc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "automatic", updatedAutomatic),
					resource.TestCheckResourceAttr(resourceName, "default_is_locked", updatedIsLocked),
					resource.TestCheckResourceAttr(resourceName, "nbd", updatedNbd)),
			},
		},
	},
	)
}

func TestAccXONetwork_updateForceNew(t *testing.T) {
	if accTestPIF.Id == "" {
		t.Skip()
	}
	resourceName := "xenorchestra_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	desc := "Non default description"
	nbd := "false"
	origMtu := "950"
	updatedMtu := "1000"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraNetworkConfigNonDefaults(nameLabel, desc, origMtu, nbd),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPIF.PoolId),
					resource.TestCheckResourceAttr(resourceName, "mtu", origMtu)),
			},
			{
				Config: testAccXenorchestraNetworkConfigNonDefaults(nameLabel, desc, updatedMtu, nbd),
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

func testAccCheckXenorchestraNetwork(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Network resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Network resource ID not set")
		}
		return nil
	}
}

var testAccXenorchestraNetworkConfig = func(name string) string {
	return fmt.Sprintf(`
resource "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}
`, name, accTestPIF.PoolId)
}

var testAccXenorchestraNetworkConfigWithoutVlan = func(name string) string {
	return testAccXenorchestraDataSourcePIFConfig(accTestPIF.Host) + fmt.Sprintf(`
resource "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
    pif_id = data.xenorchestra_pif.pif.id
}
`, name, accTestPIF.PoolId)
}

var testAccXenorchestraNetworkConfigWithoutPIF = func(name string) string {
	return fmt.Sprintf(`
resource "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
    vlan = 10
}
`, name, accTestPIF.PoolId)
}

var testAccXenorchestraNetworkConfigNonDefaults = func(name, desc, mtu, nbd string) string {
	return fmt.Sprintf(`
resource "xenorchestra_network" "network" {
    name_label = "%s"
    name_description = "%s"
    pool_id = "%s"
    mtu = %s
    nbd = %s
}
`, name, desc, accTestPIF.PoolId, mtu, nbd)
}

var testAccXenorchestraNetworkConfigNonDefaultsWithVlan = func(name, desc, mtu, nbd, vlan string) string {
	return fmt.Sprintf(`
data "xenorchestra_pif" "pif" {
    device = "eth0"
    vlan = -1
    host_id = "%s"
}

resource "xenorchestra_network" "network" {
    name_label = "%s"
    name_description = "%s"
    pool_id = "%s"
    mtu = %s
    nbd = %s
    pif_id = data.xenorchestra_pif.pif.id
    vlan = %s
}
`, accTestPIF.Host, name, desc, accTestPIF.PoolId, mtu, nbd, vlan)
}

var testAccXenorchestraNetworkConfigInPlaceUpdates = func(name, desc, nbd, automatic, isLocked string) string {
	return fmt.Sprintf(`
resource "xenorchestra_network" "network" {
    name_label = "%s"
    name_description = "%s"
    pool_id = "%s"
    nbd = %s
    automatic = %s
    default_is_locked = %s
}
`, name, desc, accTestPIF.PoolId, nbd, automatic, isLocked)
}
