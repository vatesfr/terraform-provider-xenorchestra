package xoa

import (
	"fmt"
	"log"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func createUser(t *testing.T, username string) func() {
	return func() {
		c, err := client.NewClient(client.GetConfigFromEnv())

		if err != nil {
			t.Fatalf("failed to created client with error: %v", err)
		}

		_, err = c.CreateUser(client.User{
			Email: username,
		})

		if err != nil {
			t.Fatalf("failed to create user for test with error: %v", err)
		}
	}
}

func TestAccXenorchestraDataSource_user(t *testing.T) {
	resourceName := "data.xenorchestra_user.user"
	username := fmt.Sprintf("%s-username", accTestPrefix)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: createUser(t, username),
				Config:    testAccXenorchestraDataSourceUserConfig(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceUser(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "username", username)),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceUser(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find user data source: %s", n)
		}

		log.Printf("[DEBUG] Found resource again %v", s.RootModule().Resources)
		if rs.Primary.ID == "" {
			return fmt.Errorf("User data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourceUserConfig(username string) string {
	return fmt.Sprintf(`
data "xenorchestra_user" "user" {
    username = "%s"
}
`, username)
}
