package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var createNetwork = func(net client.Network, t *testing.T, times int) func() {
	return func() {
		for i := 0; i < times; i++ {
			c, err := client.NewClient(client.GetConfigFromEnv())

			if err != nil {
				t.Fatalf("failed to created client with error: %v", err)
			}

			c.CreateNetwork(net)
		}
	}
}

var getTestNetwork = func(poolId string) client.Network {
	nameLabel := fmt.Sprintf("%s-network-%d", accTestPrefix, testObjectIndex)
	testObjectIndex++
	return client.Network{
		NameLabel: nameLabel,
		PoolId:    poolId,
	}
}

func TestAccXONetworkDataSource_read(t *testing.T) {
	resourceName := "data.xenorchestra_network.network"
	net := getTestNetwork("355ee47d-ff4c-4924-3db2-fd86ae629676")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: createNetwork(net, t, 1),
				Config:    testAccXenorchestraDataSourceNetworkConfig(net.NameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
		},
	},
	)
}

func TestAccXONetworkDataSource_multipleCauseError(t *testing.T) {
	resourceName := "data.xenorchestra_network.network"
	net := getTestNetwork("355ee47d-ff4c-4924-3db2-fd86ae629676")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: createNetwork(net, t, 2),
				Config:    testAccXenorchestraDataSourceNetworkConfig(net.NameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceNetwork(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
				ExpectError: regexp.MustCompile(`Your query returned more than one result`),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceNetwork(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Network data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Network data source ID not set")
		}
		return nil
	}
}

var testAccXenorchestraDataSourceNetworkConfig = func(name string) string {
	return fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
}
`, name)
}
