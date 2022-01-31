package xoa

import (
	"fmt"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var rsName string = fmt.Sprintf("%s-resource-set-resource", accTestPrefix)

func init() {
	resource.AddTestSweepers("xenorchestra_resource_set", &resource.Sweeper{
		Name: "xenorchestra_resource_set",
		F:    client.RemoveResourceSetsWithNamePrefix(accTestPrefix),
	})
}

func resourceSetCompositeChecks(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		testAccResourceSetExists(resourceName),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", rsName),
		resource.TestCheckResourceAttr(resourceName, "objects.#", "2"),
		resource.TestCheckResourceAttr(resourceName, "subjects.#", "2"),
		resource.TestCheckResourceAttr(resourceName, "limit.#", "2"),
		internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
			"type":     "disk",
			"quantity": "10740000000",
		}),
		internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
			"type":     "cpus",
			"quantity": "4",
		}))
}

func TestAccResourceSet_createAndPlanWithNonExistantResourceSet(t *testing.T) {
	resourceName := "xenorchestra_resource_set.bar"
	removeResourceSet := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		rs, err := c.GetResourceSet(client.ResourceSet{
			Name: rsName,
		})

		if err != nil {
			t.Fatalf("failed to find resource set with error: %v", err)
		}

		if len(rs) != 1 {
			t.Fatalf("expected to find 1 resource set, found '%d' instead", len(rs))
		}

		err = c.DeleteResourceSet(rs[0])
		if err != nil {
			t.Fatalf("failed to delete resource set with error: %v", err)
		}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResourceSetDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccResourceSetConfig(),
				Check:   resourceSetCompositeChecks(resourceName),
				Destroy: false,
			},
			{
				PreConfig:          removeResourceSet,
				Config:             testAccResourceSetConfig(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceSetExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rsName),
					resource.TestCheckResourceAttr(resourceName, "objects.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subjects.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "limit.#", "2"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
						"type":     "disk",
						"quantity": "10740000000",
					}),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
						"type":     "cpus",
						"quantity": "4",
					})),
			},
		},
	})
}

func TestAccResourceSet_addSubjectsObjectsAndLimits(t *testing.T) {
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
			{
				Config: testAccResourceSetConfigAddSubjectsObjectsAndLimits(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceSetExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rsName),
					resource.TestCheckResourceAttr(resourceName, "objects.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "subjects.#", "3"),
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
					})),
			},
		},
	})
}

func TestAccResourceSet_removeSubjectsAndObjects(t *testing.T) {
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
			{
				Config: testAccResourceSetConfigRemoveSubjectsAndObjects(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceSetExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rsName),
					resource.TestCheckResourceAttr(resourceName, "objects.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subjects.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "limit.#", "1"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "limit.*", map[string]string{
						"type":     "cpus",
						"quantity": "4",
					})),
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

    subjects = [
	"one", "two"
    ]

    objects = [
	"one", "two"
    ]
}
`, rsName)
}

func testAccResourceSetConfigAddSubjectsObjectsAndLimits() string {
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
	"one", "two", "three"
    ]

    objects = [
	"one", "two", "three"
    ]
}
`, rsName)
}

func testAccResourceSetConfigRemoveSubjectsAndObjects() string {
	return fmt.Sprintf(`
resource "xenorchestra_resource_set" "bar" {
    name = "%s"
    limit {
	type = "cpus"
	quantity = 4
    }

    subjects = [
	"one"
    ]

    objects = [
	"one"
    ]
}
`, rsName)
}
