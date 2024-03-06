package xoa

import (
	"fmt"
	"testing"

	"github.com/vatesfr/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
				Check:  aclComposeChecks(resourceName, accDefaultSr.Id, action),
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

func TestAccXenorchestraAcl_createAndRecreateOnUpdate(t *testing.T) {
	resourceName := "xenorchestra_acl.bar"
	subject := "terraform subject"
	action := "viewer"
	updatedAction := "operator"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAclConfig(subject, accDefaultSr.Id, action),
				Check:  aclComposeChecks(resourceName, accDefaultSr.Id, action),
			},
			{
				Config: testAccAclConfig(subject, accDefaultSr.Id, updatedAction),
				Check:  aclComposeChecks(resourceName, accDefaultSr.Id, updatedAction),
			},
		},
	})
}

func TestAccXenorchestraAcl_import(t *testing.T) {
	resourceName := "xenorchestra_acl.bar"
	checkFn := func(s []*terraform.InstanceState) error {
		attrs := []string{
			"id",
			"subject",
			"object",
			"action",
		}
		for _, attr := range attrs {
			_, ok := s[0].Attributes[attr]

			if !ok {
				return fmt.Errorf("attribute %s should be set", attr)
			}
		}
		return nil
	}
	subject := "terraform subject"
	action := "viewer"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAclConfig(subject, accDefaultSr.Id, action),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
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

func aclComposeChecks(resourceName, object, action string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		testAccAclExists(resourceName),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "subject"),
		resource.TestCheckResourceAttr(resourceName, "object", object),
		resource.TestCheckResourceAttr(resourceName, "action", action),
	)
}
