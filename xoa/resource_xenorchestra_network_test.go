package xoa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccXONetwork_create(t *testing.T) {
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

func TestAccXONetwork_createWithNonDefaults(t *testing.T) {
	resourceName := "xenorchestra_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	desc := "Non default description"
	// TODO(ddelnano): Add support for creating networks with VLANs
	nbd := "true"
	mtu := "950"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraNetworkConfigNonDefaults(nameLabel, desc, mtu, nbd),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPool.Id),
					resource.TestCheckResourceAttr(resourceName, "mtu", mtu),
					resource.TestCheckResourceAttr(resourceName, "nbd", nbd)),
			},
		},
	},
	)
}

func TestAccXONetwork_updateInPlace(t *testing.T) {
	resourceName := "xenorchestra_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	isLocked := "false"
	automatic := "false"
	desc := ""
	nbd := "false"

	updatedNameLabel := nameLabel + " updated"
	updatedDesc := "Non default description"
	updatedNbd := "true"
	updatedAutomatic := "true"
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
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPool.Id),
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
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPool.Id),
					resource.TestCheckResourceAttr(resourceName, "automatic", updatedAutomatic),
					resource.TestCheckResourceAttr(resourceName, "default_is_locked", updatedIsLocked),
					resource.TestCheckResourceAttr(resourceName, "nbd", updatedNbd)),
			},
		},
	},
	)
}

func TestAccXONetwork_updateForceNew(t *testing.T) {
	resourceName := "xenorchestra_network.network"
	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	desc := "Non default description"
	// TODO(ddelnano): Add support for creating networks with VLANs
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
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPool.Id),
					resource.TestCheckResourceAttr(resourceName, "mtu", origMtu)),
			},
			{
				Config: testAccXenorchestraNetworkConfigNonDefaults(nameLabel, desc, updatedMtu, nbd),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttr(resourceName, "name_description", desc),
					resource.TestCheckResourceAttr(resourceName, "pool_id", accTestPool.Id),
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
`, name, accTestPool.Id)
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
`, name, desc, accTestPool.Id, mtu, nbd)
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
`, name, desc, accTestPool.Id, nbd, automatic, isLocked)
}
