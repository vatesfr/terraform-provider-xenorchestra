package xoa

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccXenorchestraDataSource_host(t *testing.T) {
	resourceName := "data.xenorchestra_host.host"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostConfig(accTestHost.NameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceHost(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "cpus.cores"),
					resource.TestCheckResourceAttrSet(resourceName, "cpus.sockets"),
					resource.TestCheckResourceAttrSet(resourceName, "memory"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_usage"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accTestHost.NameLabel)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_hostXoTokenAuth(t *testing.T) {
	resourceName := "data.xenorchestra_host.host"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccTokenAuthProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostConfig(accTestHost.NameLabel),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceHost(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "cpus.cores"),
					resource.TestCheckResourceAttrSet(resourceName, "cpus.sockets"),
					resource.TestCheckResourceAttrSet(resourceName, "memory"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_usage"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accTestHost.NameLabel)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_hostXoTokenAuthShouldFail(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// Specify all the parameters to the provider to force the validation
				// to fail
				Config: `provider xenorchestra {
				    alias = "token_auth"
				    username = "test"
				    password = "test"
				    token = "token"
				}
				data "xenorchestra_host" "host" {
				    provider = xenorchestra.token_auth
				    name_label = "%s"
				}
				`,
				ExpectError: regexp.MustCompile(`Error: Conflicting configuration arguments`),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_hostNotFound(t *testing.T) {
	resourceName := "data.xenorchestra_host.host"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceHostConfig("not found"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceHost(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", accTestHost.NameLabel)),
				ExpectError: regexp.MustCompile(`Could not find client.Host with query`),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceHost(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Host data source: %s", n)
		}

		log.Printf("[DEBUG] Found resource again %v", s.RootModule().Resources)
		if rs.Primary.ID == "" {
			return fmt.Errorf("Host data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourceHostConfig(host string) string {
	return fmt.Sprintf(`
data "xenorchestra_host" "host" {
    name_label = "%s"
}
`, host)
}
