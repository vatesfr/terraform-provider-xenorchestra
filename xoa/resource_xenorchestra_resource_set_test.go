package xoa

import (
	"fmt"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var rsName string = "terraform-acc-resource-set-resource"

func init() {
	resource.AddTestSweepers("resource_set", &resource.Sweeper{
		Name: "resource_set",
		F:    client.RemoveResourceSetsWithNamePrefix("terraform-acc"),
	})
}

func resourceSetCompositeChecks(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		testAccResourceSetExists(resourceName),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", rsName),
		resource.TestCheckResourceAttr(resourceName, "objects.#", "2"),
		resource.TestCheckResourceAttr(resourceName, "subjects.#", "2"),
		resource.TestCheckResourceAttr(resourceName, "limit.#", "3"),
		internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
			"type":     "memory",
			"quantity": "1074000000",
		}),
		internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
			"type":     "disk",
			"quantity": "10740000000",
		}),
		internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
			"type":     "cpus",
			"quantity": "4",
		}))
}

func TestAccResourceSet_create(t *testing.T) {
	resourceName := "xenorchestra_resource_set.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig(),
				Check:  resourceSetCompositeChecks(resourceName),
			},
		},
	})
}

func TestAccResourceSet_import(t *testing.T) {
	resourceName := "xenorchestra_resource_set.bar"
	checkFn := func(s []*terraform.InstanceState) error {
		attrs := []string{
			"id",
			"name",
		}
		for _, attr := range attrs {
			_, ok := s[0].Attributes[attr]

			if !ok {
				return fmt.Errorf("attribute %s should be set", attr)
			}
		}
		return nil
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSetConfig(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
				Check:             resourceSetCompositeChecks(resourceName),
			},
		},
	})
}

func testAccResourceSetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource set Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		resourceSet, err := c.GetResourceSetById(rs.Primary.ID)

		if err != nil {
			return err
		}

		if resourceSet.Id == rs.Primary.ID {
			return nil
		}
		return nil
	}
}

func testAccCheckResourceSetDestroy(s *terraform.State) error {
	c, err := client.NewClient(client.GetConfigFromEnv())
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_resource_set" {
			continue
		}

		_, err := c.GetResourceSet(client.ResourceSet{Id: rs.Primary.ID})

		if _, ok := err.(client.NotFound); ok {
			return nil
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccResourceSetConfig() string {
	return fmt.Sprintf(`
resource "xenorchestra_resource_set" "bar" {
    name = "%s"
    limit {
	type = "cpus"
	quantity = 4
    }

    limit {
	type = "disk"
	quantity = 10740000000
    }

    limit {
	type = "memory"
	quantity = 1074000000
    }

    subjects = [
	"one", "two"
    ]

    objects = [
	"one", "two"
    ]
}
`, rsName)
}
