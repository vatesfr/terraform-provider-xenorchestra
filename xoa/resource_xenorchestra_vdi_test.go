package xoa

import (
	"errors"
	"fmt"
	"testing"

	"github.com/vatesfr/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("xenorchestra_vdi", &resource.Sweeper{
		Name: "xenorchestra_vdi",
		F:    client.RemoveVDIsWithPrefix(accTestPrefix),
	})
}

func TestAccXenorchestraVDI_readAfterDelete(t *testing.T) {
	name := fmt.Sprintf("terraform - %s", t.Name())
	resourceName := "xenorchestra_vdi.bar"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVDIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVDIConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVDIExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id")),
			},
			{
				Config:             testAccVDIConfig(name),
				Check:              testAccCheckXenorchestraVDIDestroyNow(resourceName),
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testAccVDIConfig(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccXenorchestraVDI_createAndUpdateName(t *testing.T) {
	resourceName := "xenorchestra_vdi.bar"
	name := fmt.Sprintf("terraform - %s", t.Name())
	updatedName := fmt.Sprintf("%s-updated", name)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVDIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVDIConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVDIExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "sr_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name_label"),
					resource.TestCheckResourceAttrSet(resourceName, "type"),
					resource.TestCheckResourceAttrSet(resourceName, "filepath")),
			},
			{
				Config: testAccVDIConfig(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVDIExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", fmt.Sprintf("%s%s", accTestPrefix, updatedName)),
					resource.TestCheckResourceAttrSet(resourceName, "type"),
					resource.TestCheckResourceAttrSet(resourceName, "filepath")),
			},
		},
	})
}

func testAccVDIConfig(name string) string {
	return fmt.Sprintf(`
resource "xenorchestra_vdi" "bar" {
    name_label = "%s%s"
    sr_id = "%s"
    filepath = "${path.module}/testdata/alpine-virt-3.17.0-x86_64.iso"
    type = "raw"
}
`, accTestPrefix, name, accIsoSr.Id)
}

func testAccVDIExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VDI Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		vdi, err := c.GetVDI(client.VDI{
			VDIId: rs.Primary.ID,
		})

		if vdi.VDIId == rs.Primary.ID {
			return nil
		}
		return nil
	}
}

func testAccCheckXenorchestraVDIDestroyNow(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VDI Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		err = c.DeleteVDI(rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckXenorchestraVDIDestroy(s *terraform.State) error {
	c, err := client.NewClient(client.GetConfigFromEnv())
	if err != nil {
		return err
	}
	danglingErr := errors.New("dangling VDI resource left")
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_vdi" {
			continue
		}

		_, err := c.GetVDI(client.VDI{
			VDIId: rs.Primary.ID,
		})

		if _, ok := err.(client.NotFound); ok {
			return nil
		}

		return errors.New(fmt.Sprintf("VDI with id %s left dangling", rs.Primary.ID))

	}
	return danglingErr
}
