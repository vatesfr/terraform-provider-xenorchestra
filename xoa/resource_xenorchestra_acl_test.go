package xoa

import (
	"fmt"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccXenorchestraAcl_readAfterDelete(t *testing.T) {
	resourceName := "xenorchestra_acl.bar"
	subject := "terraform subject"
	action := "viewer"
	removeAcl := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		acl, err := c.GetAcl(client.Acl{
			Subject: fmt.Sprintf("%s%s", accTestPrefix, subject),
			Object:  accDefaultSr.Id,
			Action:  action,
		})

		if err != nil {
			t.Fatalf("failed to find acl with error: %v", err)
		}

		err = c.DeleteAcl(*acl)
		if err != nil {
			t.Fatalf("failed to delete acl with error: %v", err)
		}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAclConfig(subject, accDefaultSr.Id, action),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAclExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "subject"),
					resource.TestCheckResourceAttr(resourceName, "object", accDefaultSr.Id),
					resource.TestCheckResourceAttr(resourceName, "action", action)),
			},
			{
				PreConfig:          removeAcl,
				Config:             testAccAclConfig(subject, accDefaultSr.Id, action),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccXenorchestraAcl_create(t *testing.T) {
	resourceName := "xenorchestra_acl.bar"
	subject := "terraform subject"
	action := "viewer"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAclConfig(subject, accDefaultSr.Id, action),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAclExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "subject"),
					resource.TestCheckResourceAttr(resourceName, "object", accDefaultSr.Id),
					resource.TestCheckResourceAttr(resourceName, "action", action)),
			},
		},
	})
}

// func TestAccXenorchestraAcl_import(t *testing.T) {
// 	resourceName := "xenorchestra_acl.bar"
// 	// TODO: Need to figure out how to get this to make sure all the attrs
// 	// are set. Right now it doesn't actually provide much protection
// 	checkFn := func(s []*terraform.InstanceState) error {
// 		attrs := []string{"id", "name", "template"}
// 		for _, attr := range attrs {
// 			_, ok := s[0].Attributes[attr]

// 			if !ok {
// 				return fmt.Errorf("attribute %s should be set", attr)
// 			}
// 		}
// 		return nil
// 	}
// 	templateName := "testing"
// 	templateText := "template body"
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckXenorchestraAclDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAclConfig(templateName, templateText),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateCheck:  checkFn,
// 				ImportStateVerify: true,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					testAccAclExists(resourceName),
// 					resource.TestCheckResourceAttrSet(resourceName, "id")),
// 			},
// 		},
// 	})
// }

func testAccAclConfig(subject, object, action string) string {
	return fmt.Sprintf(`
resource "xenorchestra_acl" "bar" {
    subject = "%s%s"
    object = "%s"
    action = "%s"
}
`, accTestPrefix, subject, object, action)
}

func testAccAclExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Acl Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		acl, err := c.GetAcl(client.Acl{Id: rs.Primary.ID})

		if acl.Id == rs.Primary.ID {
			return nil
		}
		return nil
	}
}

func testAccCheckXenorchestraAclDestroyNow(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Acl Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		err = c.DeleteAcl(client.Acl{Id: rs.Primary.ID})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckXenorchestraAclDestroy(s *terraform.State) error {
	c, err := client.NewClient(client.GetConfigFromEnv())
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_acl" {
			continue
		}

		acl, err := c.GetAcl(client.Acl{Id: rs.Primary.ID})

		if _, ok := err.(client.NotFound); ok {
			return nil
		}

		if err != nil {
			return err
		}

		if acl != nil {
			return fmt.Errorf("Acl (%s) still exists", acl.Id)
		}
	}
	return nil
}
