package xoa

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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

func TestAccXenorchestraDataSource_userInCurrentSession(t *testing.T) {
	resourceName := "data.xenorchestra_user.user"
	username := fmt.Sprintf("%s-username2", accTestPrefix)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: createUser(t, username),
				Config:    testAccXenorchestraDataSourceUserInCurrentSessionConfig(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceUser(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "username", username)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_userNotFound(t *testing.T) {
	resourceName := "data.xenorchestra_user.user"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceUserConfig("not found"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceUser(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
				ExpectError: regexp.MustCompile(`Could not find client.User with query`),
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

func testAccXenorchestraDataSourceUserInCurrentSessionConfig(username string) string {
	return fmt.Sprintf(`
provider "xenorchestra" {
    username = "%s"
    password = "password"
}

data "xenorchestra_user" "user" {
    username = "%s"
    search_in_session = true
}
`, username, username)
}

func createUser(t *testing.T, username string) func() {
	return func() {
		c, err := client.NewClient(client.GetConfigFromEnv())

		if err != nil {
			t.Fatalf("failed to created client with error: %v", err)
		}

		_, err = c.CreateUser(client.User{
			Email:    username,
			Password: "password",
		})

		if err != nil {
			t.Fatalf("failed to create user for test with error: %v", err)
		}
	}
}
