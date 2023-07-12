package xoa

import (
	"fmt"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var netName string = fmt.Sprintf("%s-network-resource", accTestPrefix)

func init() {
	resource.AddTestSweepers("xenorchestra_network", &resource.Sweeper{
		Name: "xenorchestra_network",
		F:    client.RemoveResourceSetsWithNamePrefix(accTestPrefix),
	})
}

func TestAccXenorchestraNetwork_readAfterDelete(t *testing.T) {
	networkName := "testNetwork"
	poolId := accTestPool.Id
	resourceName := "xenorchestra_network.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenOrchestraNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetwork(networkName, poolId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCloudConfigExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
			{
				Config:             testAccNetwork(networkName, poolId),
				Check:              testAccCheckXenOrchestraNetworkDestroyNow(resourceName),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccNetwork(networkName, poolId),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccXenorchestraNetwork_create(t *testing.T) {

}

func TestAccXenorchestraNetwork_update(t *testing.T) {

}

func testAccNetwork(name, poolId string) string {
	return fmt.Sprintf(`
resource "xenorchestra_network" "bar" {
	name_label = "%s%s"
	pool_id = "%s"
	description = "Acceptance test network"
	pif = "TODO: PIF"
	mtu = 1500
	vlan = 100
}`, accTestPrefix, name, poolId)
}

func testAccNetworkExists(resourceName string) resource.TestCheckFunc {
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

		net, err := c.GetNetwork(client.Network{Id: rs.Primary.ID})

		if err != nil {
			return err
		}

		if net.Id == rs.Primary.ID {
			return nil
		}

		return nil
	}
}

func testAccCheckXenOrchestraNetworkDestroy(s *terraform.State) error {
	c, err := client.NewClient(client.GetConfigFromEnv())

	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_network" {
			continue
		}

		net, err := c.GetNetwork(client.Network{Id: rs.Primary.ID})

		if err != nil {
			return err
		}

		if net != nil {
			return fmt.Errorf("Network (%s) still exists", net.Id)
		}
	}

	return nil
}

func testAccCheckXenOrchestraNetworkDestroyNow(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		err = c.DeleteNetwork(rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}
