package xoa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testResourceSetName string = "terraform-acc-data-source-test"

var getTestResourceSet = func(name string) client.ResourceSet {
	return client.ResourceSet{
		Name: name,
		Limits: client.ResourceSetLimits{
			Cpus: client.ResourceSetLimit{
				Total:     1,
				Available: 2,
			},
			Disk: client.ResourceSetLimit{
				Total:     1,
				Available: 2,
			},
			Memory: client.ResourceSetLimit{
				Total:     1,
				Available: 2,
			},
		},
		Subjects: []string{},
		Objects:  []string{},
	}
}

var rsIdx int = 1

func getUniqResourceSetName(t *testing.T) string {
	rv := fmt.Sprintf("%s%d - %s", testResourceSetName, rsIdx, t.Name())
	rsIdx++
	return rv
}

var createResourceSet = func(t *testing.T, name string, count int) func() {
	return func() {
		c, err := client.NewClient(client.GetConfigFromEnv())

		if err != nil {
			t.Fatalf("failed to created client with error: %v", err)
		}

		for i := 0; i < count; i++ {
			_, err = c.CreateResourceSet(
				getTestResourceSet(name),
			)

			if err != nil {
				t.Fatalf("failed to created resource set with error: %v", err)
			}
		}
	}
}

func TestAccXenorchestraDataSource_resourceSet(t *testing.T) {
	resourceName := "data.xenorchestra_resource_set.rs"
	rsName := getUniqResourceSetName(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: createResourceSet(t, rsName, 1),
				Config:    testAccXenorchestraDataSourceResourceSetConfig(rsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceResourceSet(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rsName)),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_resourceSetNotFound(t *testing.T) {
	resourceName := "data.xenorchestra_resource_set.rs"
	rsName := getUniqResourceSetName(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccXenorchestraDataSourceResourceSetConfig("not found"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXenorchestraDataSourceResourceSet(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rsName)),
				ExpectError: regexp.MustCompile(`Could not find client.ResourceSet with query`),
			},
		},
	},
	)
}

func TestAccXenorchestraDataSource_withDuplicateResourceSetNames(t *testing.T) {
	rsName := getUniqResourceSetName(t)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig:   createResourceSet(t, rsName, 2),
				Config:      testAccXenorchestraDataSourceResourceSetConfig(rsName),
				ExpectError: regexp.MustCompile("Resource sets must be uniquely named to use this data source"),
			},
		},
	},
	)
}

func testAccCheckXenorchestraDataSourceResourceSet(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find ResourceSet data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ResourceSet data source ID not set")
		}
		return nil
	}
}

func testAccXenorchestraDataSourceResourceSetConfig(name string) string {
	return fmt.Sprintf(`
data "xenorchestra_resource_set" "rs" {
    name = "%s"
}
`, name)
}
