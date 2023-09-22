package xoa

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("xenorchestra_vm", &resource.Sweeper{
		Name:         "xenorchestra_vm",
		F:            client.RemoveVmsWithNamePrefix(accTestPrefix),
		Dependencies: []string{"xenorchestra_resource_set", "xenorchestra_cloud_config"},
	})
}

func Test_extractIpsFromNetworks(t *testing.T) {
	ipv4 := "169.254.169.254"
	secondIpv4 := "169.254.255.254"
	ipv6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	secondIpv6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7000"
	tests := []struct {
		networks map[string]string
		expected []map[string][]string
	}{
		{
			networks: map[string]string{},
			expected: []map[string][]string{},
		},
		{
			networks: map[string]string{
				"0/ip":     ipv4,
				"0/ipv4/0": ipv4,
				"0/ipv4/1": secondIpv4,
				"0/ipv6/0": ipv6,
				"0/ipv6/1": secondIpv6,
				"1/ip":     ipv4,
				"1/ipv4/0": ipv4,
				"1/ipv4/1": secondIpv4,
				"1/ipv6/0": ipv6,
				"1/ipv6/1": secondIpv6,
			},
			expected: []map[string][]string{
				map[string][]string{
					"ip":   []string{ipv4},
					"ipv4": []string{ipv4, secondIpv4},
					"ipv6": []string{ipv6, secondIpv6},
				},
				map[string][]string{
					"ip":   []string{ipv4},
					"ipv4": []string{ipv4, secondIpv4},
					"ipv6": []string{ipv6, secondIpv6},
				},
			},
		},
	}

	for _, test := range tests {
		expected := test.expected
		nets := test.networks
		actual := extractIpsFromNetworks(nets)

		if len(expected) != len(actual) {
			t.Errorf("expected '%+v' to have the same length as: %+v", expected, actual)
		}

		for device := 0; device < len(expected); device++ {
			for _, key := range []string{"ip", "ipv4", "ipv6"} {
				if !reflect.DeepEqual(expected[device][key], actual[device][key]) {
					t.Errorf("expected '%+v' to be equal to: %+v", expected[device][key], actual[device][key])
				}
			}
		}
	}
}

func Test_diskHash(t *testing.T) {
	nameLabel := "name label"
	nameDescription := "name description"
	attached := true
	size := 1000
	srId := "sr id"
	cases := []struct {
		clientDisk client.Disk
		mapDisk    map[string]interface{}
	}{
		{
			clientDisk: client.Disk{
				client.VBD{
					Attached: attached,
				},
				client.VDI{
					NameLabel:       nameLabel,
					NameDescription: nameDescription,
					SrId:            srId,
					Size:            size,
				},
			},
			mapDisk: map[string]interface{}{
				"name_label":       nameLabel,
				"name_description": nameDescription,
				"attached":         attached,
				"sr_id":            srId,
				"size":             size,
			},
		},
	}

	for _, c := range cases {
		cDiskHash := diskHash(c.clientDisk)
		mapDiskHash := diskHash(c.mapDisk)
		if cDiskHash != mapDiskHash {
			t.Errorf("expected the hash of %+v to match the disk map: %+v. instead received %d and %d", c.clientDisk, c.mapDisk, cDiskHash, mapDiskHash)
		}
	}
}

func Test_getUpdateDiskActions(t *testing.T) {
	cases := []struct {
		disk                client.Disk
		haystack            []client.Disk
		expectedDiskActions []updateDiskActions
	}{
		{
			disk: client.Disk{
				client.VBD{
					Id:       "id 1",
					Attached: true,
				},
				client.VDI{},
			},
			haystack: []client.Disk{
				{
					client.VBD{
						Id:       "id 1",
						Attached: false,
					},
					client.VDI{},
				},
			},
			expectedDiskActions: []updateDiskActions{diskAttachmentUpdate},
		},
		{
			disk: client.Disk{
				client.VBD{
					Id:       "id 1",
					Attached: true,
				},
				client.VDI{
					NameLabel:       "name label",
					NameDescription: "name description",
				},
			},
			haystack: []client.Disk{
				{
					client.VBD{
						Id:       "id 1",
						Attached: false,
					},
					client.VDI{
						NameLabel:       "updated name label",
						NameDescription: "updated name description",
					},
				},
			},
			expectedDiskActions: []updateDiskActions{diskNameLabelUpdate, diskNameDescriptionUpdate, diskAttachmentUpdate},
		},
		{
			disk: client.Disk{
				client.VBD{
					Id:       "does not match",
					Attached: true,
				},
				client.VDI{},
			},
			haystack: []client.Disk{
				{
					client.VBD{
						Id:       "id 1",
						Attached: false,
					},
					client.VDI{},
				},
			},
			expectedDiskActions: []updateDiskActions{},
		},
	}

	for _, c := range cases {
		actions := getUpdateDiskActions(c.disk, c.haystack)

		if !reflect.DeepEqual(c.expectedDiskActions, actions) {
			t.Errorf("expected updateDiskActions '%+v' to match '%+v' when comparing disk: %+v against the following disks: %+v", c.expectedDiskActions, actions, c.disk, c.haystack)
		}
	}
}

func Test_shouldUpdateDisk(t *testing.T) {
	cases := []struct {
		disk                 client.Disk
		haystack             []client.Disk
		expectedShouldUpdate bool
	}{
		{
			disk: client.Disk{
				client.VBD{
					Id:       "id 1",
					Attached: true,
				},
				client.VDI{},
			},
			haystack: []client.Disk{
				{
					client.VBD{
						Id:       "id 1",
						Attached: false,
					},
					client.VDI{},
				},
			},
			expectedShouldUpdate: true,
		},
		{
			disk: client.Disk{
				client.VBD{
					Id:       "does not match",
					Attached: true,
				},
				client.VDI{},
			},
			haystack: []client.Disk{
				{
					client.VBD{
						Id:       "id 1",
						Attached: false,
					},
					client.VDI{},
				},
			},
			expectedShouldUpdate: false,
		},
	}

	for _, c := range cases {
		shouldUpdate := shouldUpdateDisk(c.disk, c.haystack)

		if c.expectedShouldUpdate != shouldUpdate {
			t.Errorf("expected shouldUpdate '%t' to match '%t' when comparing disk: %+v against the following disks: %+v", c.expectedShouldUpdate, shouldUpdate, c.disk, c.haystack)
		}
	}
}

func Test_shouldUpdateVif(t *testing.T) {
	cases := []struct {
		vif                  client.VIF
		haystack             []*client.VIF
		expectedShouldUpdate bool
		expectedShouldAttach bool
	}{
		{
			vif: client.VIF{
				MacAddress: "mac address",
				Attached:   true,
			},
			haystack: []*client.VIF{
				&client.VIF{
					Id:         "id",
					MacAddress: "mac address",
					Attached:   false,
				},
			},
			expectedShouldUpdate: true,
			expectedShouldAttach: true,
		},
		{
			vif: client.VIF{
				Id:       "id",
				Attached: true,
			},
			haystack: []*client.VIF{
				&client.VIF{
					Id:       "id",
					Attached: false,
				},
			},
			expectedShouldUpdate: true,
			expectedShouldAttach: true,
		},
		{
			vif: client.VIF{
				Id:       "id",
				Attached: false,
			},
			haystack: []*client.VIF{
				&client.VIF{
					Id:       "id",
					Attached: false,
				},
			},
			expectedShouldUpdate: false,
			expectedShouldAttach: false,
		},
	}

	for _, c := range cases {
		shouldUpdate, shouldAttach := shouldUpdateVif(c.vif, c.haystack)

		if c.expectedShouldUpdate != shouldUpdate {
			t.Errorf("expected shouldUpdate '%t' to match '%t' when comparing VIF: %+v against the following VIFs: %+v", c.expectedShouldUpdate, shouldUpdate, c.vif, c.haystack)
		}

		if c.expectedShouldAttach != shouldAttach {
			t.Errorf("expected shouldAttach '%t' to match '%t' when comparing VIF: %+v against the following VIFs: %+v", c.expectedShouldAttach, shouldAttach, c.vif, c.haystack)
		}
	}
}

func TestAccXenorchestraVm_createWithShorterResourceTimeout(t *testing.T) {
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVmConfigWithShortTimeout(vmName),
				ExpectError: regexp.MustCompile("timeout while waiting for state to become"),
			},
		},
	})
}

func TestAccXenorchestraVm_createAndPlanWithNonExistantVm(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	removeVm := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		vm, err := c.GetVm(client.Vm{
			NameLabel: vmName,
		})

		if err != nil {
			t.Fatalf("failed to find VM with error: %v", err)
		}

		err = c.DeleteVm(vm.Id)
		if err != nil {
			t.Fatalf("failed to delete VM with error: %v", err)
		}
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
				Destroy: false,
			},
			{
				PreConfig:          removeVm,
				Config:             testAccVmConfig(vmName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccXenorchestraVm_createWithDestroyCloudConfigDrive(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	verifyCloudConfigDiskDeleted := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		vm, err := c.GetVm(client.Vm{
			NameLabel: vmName,
		})

		if err != nil {
			t.Fatalf("failed to find VM with error: %v", err)
		}

		vmDisks, err := c.GetDisks(vm)
		if err != nil {
			t.Fatalf("failed to get Vm's disks with error: %v", err)
		}

		for _, disk := range vmDisks {
			if disk.NameLabel == defaultCloudConfigDiskName {
				t.Errorf("expected the VM to have its cloud config VDI removed, instead found: %v", disk)
			}
		}
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithDestroyCloudConfigAfterBoot(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
				Destroy: false,
			},
			{
				PreConfig: verifyCloudConfigDiskDeleted,
				Config:    testAccVmConfigWithDestroyCloudConfigAfterBoot(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
				PlanOnly: true,
			},
		},
	})
}

func TestAccXenorchestraVm_createWhenWaitingForIp(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	regex := regexp.MustCompile(`[1-9]*`)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWaitForIp(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_ip", "true"),
					resource.TestMatchResourceAttr(resourceName, "ipv6_addresses.#", regex),
					resource.TestCheckResourceAttrSet(resourceName, "ipv6_addresses.0"),
					resource.TestMatchResourceAttr(resourceName, "network.0.ipv6_addresses.#", regex),
					resource.TestCheckResourceAttrSet(resourceName, "network.0.ipv6_addresses.0"),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_ensureVmsInResourceSetsCanBeUpdatedByNonAdminUsers(t *testing.T) {
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	adminUser := os.Getenv("XOA_USER")
	adminPassword := os.Getenv("XOA_PASSWORD")
	accUserPassword := "password"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			// Create a resource set and cloud config template with an admin user
			{
				Config: testAccVmResourceSet(vmName) + testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template"),
			},
			// Create a VM using the resource set from the previous step
			{
				Config: providerCredentials(accUser.Email, accUserPassword) +
					testAccVmManagedResourceSetConfig(vmName),
			},
			// Verify that the non admin user can update the VM. This is the main assertion of the test
			{
				Config: providerCredentials(accUser.Email, accUserPassword) +
					testAccVmManagedResourceSetWithDescriptionConfig(vmName, "new description"),
			},
			// Re-run with the admin user so that it can delete the resource set and cloud config
			{
				Config: providerCredentials(adminUser, adminPassword) +
					testAccVmManagedResourceSetConfig(vmName),
			},
		},
	})
}

func TestAccXenorchestraVm_cdromAndInstallationMethodsCannotBeSpecifiedTogether(t *testing.T) {
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVmConfigConflictingCdromAndInstallMethod(vmName),
				ExpectError: regexp.MustCompile(`"installation_method": conflicts with cdrom`),
			},
		},
	})
}

func TestAccXenorchestraVm_createVmThatInstallsFromTheNetwork(t *testing.T) {
	t.Skip("For now this test is not implemented. See #156 for more details")
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigPXEBoot(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "installation_method", "network"),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_createAndUpdateDiskNameLabelAndNameDescription(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	nameLabel := "disk name label"
	description := "disk description"
	updatedNameLabel := "updated disk name label"
	updatedDescription := "updated description"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithDiskNameLabelAndNameDescription(vmName, nameLabel, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "memory_max", "4295000000"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.name_label", nameLabel),
					resource.TestCheckResourceAttr(resourceName, "disk.0.name_description", description)),
			},
			{
				Config: testAccVmConfigWithDiskNameLabelAndNameDescription(vmName, updatedNameLabel, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.name_label", updatedNameLabel),
					resource.TestCheckResourceAttr(resourceName, "disk.0.name_description", updatedDescription)),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithTags(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	tag1 := "tag1"
	tag2 := "tag2"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithTags(vmName, tag1, tag2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					internal.TestCheckTypeSetAttr(resourceName, "tags.*", tag1),
					internal.TestCheckTypeSetAttr(resourceName, "tags.*", tag2),
				),
			},
			{
				Config:   testAccVmConfigWithTags(vmName, tag2, tag1),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					internal.TestCheckTypeSetAttr(resourceName, "tags.*", tag1),
					internal.TestCheckTypeSetAttr(resourceName, "tags.*", tag2),
				),
			},
			{
				Config: testAccVmConfigWithTag(vmName, tag1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					internal.TestCheckTypeSetAttr(resourceName, "tags.*", tag1),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithDisklessTemplateAndISO(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithISO(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cdrom.#", "1"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "cdrom.0.*", "data.xenorchestra_vdi.iso", "id"),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_insertAndEjectCd(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccVmConfigWithCd(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cdrom.#", "1"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "cdrom.0.*", "data.xenorchestra_vdi.iso", "id"),
				),
			},
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cdrom.#", "0"),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithAffinityHost(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	affinityHost := accTestPool.Master
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithAffinityHost(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "affinity_host", affinityHost)),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithoutCloudConfig(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmWithoutCloudInitConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithCloudInitNetworkConfig(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithNetworkConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "cloud_network_config", regexp.MustCompile("type: physical")),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithDashedMacAddress(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	macWithDashes := "00-0a-83-b1-c0-01"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithMacAddress(vmName, macWithDashes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "network.*", map[string]string{
						// All mac addresses should be formatted to use colons
						"mac_address": getFormattedMac(macWithDashes),
					}),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_createAndUpdateWithMacAddress(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	macAddress := "00:0a:83:b1:c0:83"
	otherMacAddress := "00:0a:83:b1:c0:00"
	macWithDashes := "00-0a-83-b1-c0-01"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithMacAddress(vmName, macAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "network.*", map[string]string{
						"mac_address": macAddress,
					}),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config: testAccVmConfigWithMacAddress(vmName, otherMacAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "network.*", map[string]string{
						"mac_address": otherMacAddress,
					}),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config: testAccVmConfigWithMacAddress(vmName, macWithDashes),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					internal.TestCheckTypeSetElemNestedAttrs(resourceName, "network.*", map[string]string{
						// All mac addresses should be formatted to use colons
						"mac_address": getFormattedMac(macWithDashes),
					}),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_disconnectAttachedVif(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmVifAttachedConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config:             testAccVmVifDetachedConfig(vmName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVmVifDetachedConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "false"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_attachDisconnectedVif(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	removeVif := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		vm, err := c.GetVm(client.Vm{
			NameLabel: vmName,
		})

		if err != nil {
			t.Fatalf("failed to find VM with error: %v", err)
		}

		// Sleep so that the VM has a change to load the PV drivers
		time.Sleep(20 * time.Second)
		err = c.DisconnectVIF(&client.VIF{Id: vm.VIFs[0]})
		if err != nil {
			t.Fatalf("failed to disconnect VIF with error: %v", err)
		}
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmVifAttachedConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				PreConfig:          removeVif,
				Config:             testAccVmVifAttachedConfig(vmName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVmVifAttachedConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "true"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_attachDisconnectedDisk(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	disconnectDisk := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		vm, err := c.GetVm(client.Vm{
			NameLabel: vmName,
		})

		if err != nil {
			t.Fatalf("failed to find VM with error: %v", err)
		}

		disks, err := c.GetDisks(vm)

		// Sleep so that the VM has a change to load the PV drivers
		time.Sleep(120 * time.Second)

		if err != nil {
			t.Fatalf("failed to retrieve the following vm's disks: %+v with error: %v", vm, err)
		}

		for _, disk := range disks {
			if disk.NameLabel != "disk 2" {
				continue
			}
			err = c.DisconnectDisk(disk)
			if err != nil {
				t.Fatalf("failed to disconnect disk with error: %v", err)
			}
		}
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithAdditionalDisk(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.attached", "true"),
					resource.TestCheckResourceAttr(resourceName, "disk.1.attached", "true")),
			},
			{
				PreConfig: disconnectDisk,
				Config:    testAccVmConfigWithAdditionalDisk(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.attached", "true"),
					resource.TestCheckResourceAttr(resourceName, "disk.1.attached", "false")),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVmConfigWithAdditionalDisk(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.attached", "true"),
					resource.TestCheckResourceAttr(resourceName, "disk.1.attached", "true")),
			},
		},
	})
}

func TestAccXenorchestraVm_disconnectAttachedDisk(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.attached", "true")),
			},
			{
				Config:             testAccVmConfigDisconnectedDisk(vmName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVmConfigDisconnectedDisk(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.attached", "false")),
			},
		},
	})
}

func TestAccXenorchestraVm_createWithMutipleDisks(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithAdditionalDisk(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "disk.0.attached", "true"),
					resource.TestCheckResourceAttr(resourceName, "disk.1.attached", "true")),
			},
		},
	})
}

func TestAccXenorchestraVm_addAndRemoveDisksToVm(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "1"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				PreConfig: func() {
					time.Sleep(20 * time.Second)
				},
				Config: testAccVmConfigWithAdditionalDisk(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "2"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				PreConfig: func() {
					time.Sleep(20 * time.Second)
				},
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "disk.#", "1"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_import(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	checkFn := func(s []*terraform.InstanceState) error {
		attrs := []string{"id", "name_label"}
		for _, attr := range attrs {
			_, ok := s[0].Attributes[attr]

			if !ok {
				return fmt.Errorf("attribute %s should be set", attr)
			}
		}
		return nil
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(vmName),
			},
			{
				ResourceName:     resourceName,
				ImportState:      true,
				ImportStateCheck: checkFn,
				// TODO: Need to store all the
				// schema.Schema structs in the statefile that
				// currently exist before this will pass.
				// ImportStateVerify: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_description", "description"),
					resource.TestCheckResourceAttr(resourceName, "name_label", vmName),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

// TODO: Add unit tests
func testAccCheckXenorchestraVmDestroy(s *terraform.State) error {
	c, err := client.NewClient(client.GetConfigFromEnv())
	if err != nil {
		return err
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "xenorchestra_vm" {
			continue
		}

		_, err := c.GetVm(client.Vm{Id: rs.Primary.ID})

		if _, ok := err.(client.NotFound); ok {
			return nil
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func TestAccXenorchestraVm_addVifAndRemoveVif(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config: testAccVmConfigWithSecondVIF(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "2"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.0.*", "data.xenorchestra_network.network", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.device", "0"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "true"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.1.*", "data.xenorchestra_network.network2", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.1.device", "1"),
					resource.TestCheckResourceAttr(resourceName, "network.1.attached", "true")),
			},
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.0.*", "data.xenorchestra_network.network", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.device", "0"),
					resource.TestCheckResourceAttr(resourceName, "network.0.attached", "true")),
			},
		},
	})
}

func TestAccXenorchestraVm_replaceExistingVifs(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	firstMacAddress := "02:00:00:00:00:00"
	secondMacAddress := "02:00:00:00:00:11"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigWithTwoMacAddresses(vmName, firstMacAddress, secondMacAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.0.mac_address", firstMacAddress),
					resource.TestCheckResourceAttr(resourceName, "network.1.mac_address", secondMacAddress)),
			},
			{
				Config: testAccVmConfigWithTwoMacAddresses(vmName, secondMacAddress, firstMacAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network.#", "2"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.0.*", "data.xenorchestra_network.network", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.mac_address", secondMacAddress),
					resource.TestCheckResourceAttr(resourceName, "network.1.mac_address", firstMacAddress)),
			},
		},
	})
}

func TestAccXenorchestraVm_updatesWithoutReboot(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"

	origNameLabel := fmt.Sprintf("%s - orig label (%s)", accTestPrefix, t.Name())
	origNameDesc := "name label"
	origHa := "restart"
	origPowerOn := true
	updatedNameLabel := fmt.Sprintf("%s - updated label (%s)", accTestPrefix, t.Name())
	updatedNameDesc := "Terraform Updated description"
	updatedHa := ""
	updatedPowerOn := false
	affinityHost := accTestPool.Master
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigUpdateAttrsHaltIrrelevant(origNameLabel, origNameDesc, origHa, origPowerOn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", origNameLabel),
					resource.TestCheckResourceAttr(resourceName, "auto_poweron", strconv.FormatBool(origPowerOn)),
					resource.TestCheckResourceAttr(resourceName, "high_availability", origHa)),
			},
			{
				Config: testAccVmConfigUpdateAttrsHaltIrrelevantWithAffinityHost(updatedNameLabel, updatedNameDesc, updatedHa, updatedPowerOn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name_label", updatedNameLabel),
					resource.TestCheckResourceAttr(resourceName, "name_description", updatedNameDesc),
					resource.TestCheckResourceAttr(resourceName, "affinity_host", affinityHost),
					resource.TestCheckResourceAttr(resourceName, "auto_poweron", strconv.FormatBool(updatedPowerOn)),
					resource.TestCheckResourceAttr(resourceName, "high_availability", updatedHa)),
			},
		},
	})
}

func TestAccXenorchestraVm_updatesWithoutRebootForOtherAttrs(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"

	nameLabel := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigUpdateAttr(
					nameLabel,
					"",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: testAccVmConfigUpdateAttr(
					nameLabel,
					`
                                    exp_nested_hvm = true
                            `),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "exp_nested_hvm", "true"),
				),
			},
			{
				Config: testAccVmConfigUpdateAttr(
					nameLabel,
					`
                                    hvm_boot_firmware = "uefi"
                            `),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "hvm_boot_firmware", "uefi"),
				),
			},
			// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
			// {
			// 	Config: testAccVmConfigUpdateAttr(
			// 		nameLabel,
			// 		`
			// hvm_boot_firmware = "uefi"
			// secure_boot = true
			// `),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccVmExists(resourceName),
			// 		resource.TestCheckResourceAttrSet(resourceName, "id"),
			// 		resource.TestCheckResourceAttr(resourceName, "hvm_boot_firmware", "uefi"),
			// 		resource.TestCheckResourceAttr(resourceName, "secure_boot", "true"),
			// 	),
			// },

			{
				Config: testAccVmConfigUpdateAttr(
					nameLabel,
					`
                              blocked_operations = ["copy"]
                            `),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "blocked_operations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "blocked_operations.0", "copy"),
				),
			},

			{
				Config: testAccVmConfigUpdateAttr(
					nameLabel,
					`
                                    vga = "cirrus"
                                    videoram = 16
                            `),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "vga", "cirrus"),
					resource.TestCheckResourceAttr(resourceName, "videoram", "16"),
				),
			},
			{
				Config: testAccVmConfigUpdateAttr(
					nameLabel,
					`
				    start_delay = 1
			`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "start_delay", "1"),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_updatesThatRequireReboot(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigUpdateAttrsVariableCPUAndMemory(2, 4295000000, vmName, "", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory_max", "4295000000"),
				),
			},
			{
				Config: testAccVmConfigUpdateAttrsVariableCPUAndMemory(5, 6295000000, vmName, "", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cpus", "5"),
					resource.TestCheckResourceAttr(resourceName, "memory_max", "6295000000"),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_updatingCpusInsideMaxCpuAndMemInsideStaticMaxDoesNotRequireReboot(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		// Use a provider that has a XO client that will error if StartVm
		// or HaltVm are called. This ensures that the VM is not rebooted during
		// the test to prove that the CPUs are changed online
		Providers:    testAccFailToStartAndHaltProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfigUpdateAttrsVariableCPUAndMemory(5, 4295000000, vmName, "", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cpus", "5"),
					resource.TestCheckResourceAttr(resourceName, "memory_max", "4295000000"),
				),
			},
			{
				Config: testAccVmConfigUpdateAttrsVariableCPUAndMemory(2, 3221225472, vmName, "", "", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "memory_max", "3221225472"),
				),
			},
		},
	})
}

func TestAccXenorchestraVm_createAndUpdateWithResourceSet(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmManagedResourceSetConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_set"),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
			{
				Config: testAccVmConfigWithoutResourceSet(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "resource_set", ""),
					internal.TestCheckTypeSetElemAttrPair(resourceName, "network.*.*", "data.xenorchestra_network.network", "id")),
			},
		},
	})
}

func TestAccXenorchestraVm_diskAndNetworkAttachmentIgnoredWhenHalted(t *testing.T) {
	resourceName := "xenorchestra_vm.bar"
	vmName := fmt.Sprintf("%s - %s", accTestPrefix, t.Name())
	shutdownVm := func() {
		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			t.Fatalf("failed to create client with error: %v", err)
		}

		vm, err := c.GetVm(client.Vm{
			NameLabel: vmName,
		})

		err = c.HaltVm(vm.Id)

		if err != nil {
			t.Fatalf("failed to halt VM with error: %v", err)
		}
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckXenorchestraVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVmConfig(vmName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
				),
			},
			{
				PreConfig: shutdownVm,
				Config:    testAccVmConfig(vmName),
				PlanOnly:  true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVmExists(resourceName),
				),
			},
		},
	})
}

func testAccVmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Vm Id is set")
		}

		c, err := client.NewClient(client.GetConfigFromEnv())
		if err != nil {
			return err
		}

		vm, err := c.GetVm(client.Vm{Id: rs.Primary.ID})

		if err != nil {
			return err
		}

		if vm.Id == rs.Primary.ID {
			return nil
		}
		return nil
	}
}

func testAccVmWithoutCloudInitConfig(vmName string) string {
	return testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithTag(vmName, tag string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }

    tags = [
      "%s",
    ]
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id, tag)
}

func testAccVmConfigWithISO(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccNonDefaultTemplateConfig(disklessTestTemplate.NameLabel) + fmt.Sprintf(`
data "xenorchestra_vdi" "iso" {
    name_label = "%s"
    pool_id = "%s"
}

data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    cdrom {
      id = data.xenorchestra_vdi.iso.id
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, testIsoName, accTestPool.Id, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithoutISO(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccNonDefaultTemplateConfig(disklessTestTemplate.NameLabel) + fmt.Sprintf(`

data "xenorchestra_vdi" "iso" {
    name_label = "%s"
    pool_id = "%s"
}

data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, testIsoName, accTestPool.Id, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithTags(vmName, tag, secondTag string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }

    tags = [
      "%s",
      "%s",
    ]
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id, tag, secondTag)
}

func testAccVmConfigWithAffinityHost(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_pool" "pool" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "${data.xenorchestra_pool.pool.id}"
}

resource "xenorchestra_vm" "bar" {
    affinity_host = "${data.xenorchestra_pool.pool.master}"
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accTestPool.NameLabel, accDefaultNetwork.NameLabel, vmName, accDefaultSr.Id)
}

func testAccVmConfig(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

// This sets destroy_cloud_config_vdi_after_boot and wait_for_ip. The former is required for
// the test expectations while the latter is to ensure the test holds its assertions until the
// disk was actually deleted. The XO api uses the guest metrics to determine when it can remove
// the disk, so an IP address allocation happens at the same time.
func testAccVmConfigWithDestroyCloudConfigAfterBoot(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    destroy_cloud_config_vdi_after_boot = true
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }
    wait_for_ip = true

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigPXEBoot(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccNonDefaultTemplateConfig(disklessTestTemplate.NameLabel) + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "Lab Network (VLAN 10)"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }
    installation_method = "network"

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 20001317888
    }
}
`, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigConflictingCdromAndInstallMethod(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_vdi" "iso" {
    name_label = "%s"
    pool_id = "%s"
}

data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }
    cdrom {
	id = data.xenorchestra_vdi.iso.id
    }
    installation_method = "network"

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, testIsoName, accTestPool.Id, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithShortTimeout(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }

    timeouts {
	create = "5s"
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithCd(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_vdi" "iso" {
    name_label = "%s"
    pool_id = "%s"
}

data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    cdrom {
	id = data.xenorchestra_vdi.iso.id
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, testIsoName, accTestPool.Id, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWaitForIp(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    wait_for_ip = true
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithDiskNameLabelAndNameDescription(vmName, nameLabel, description string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "%s"
      name_description = "%s"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id, nameLabel, description)
}

func testAccVmConfigWithNetworkConfig(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    cloud_network_config = <<EOF
    network:
      version: 1
      config:
        # Physical interfaces.
        - type: physical
          name: eth0
          mac_address: c0:d6:9f:2c:e8:80
EOF
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigDisconnectedDisk(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
      attached = false
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithAdditionalDisk(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }

    disk {
      sr_id = "%s"
      name_label = "disk 2"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id, accDefaultSr.Id)
}

func testAccVmVifAttachedConfig(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
	attached = true
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmVifDetachedConfig(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
	attached = false
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithMacAddress(vmName, macAddress string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
	mac_address = "%s"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, macAddress, accDefaultSr.Id)
}

func testAccVmConfigWithTwoMacAddresses(vmName, firstMac, secondMac string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
	mac_address = "%s"
    }

    network {
	network_id = "${data.xenorchestra_network.network.id}"
	mac_address = "%s"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, firstMac, secondMac, accDefaultSr.Id)
}

func testAccVmConfigWithSecondVIF(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

data "xenorchestra_network" "network2" {
    name_label = "Pool-wide network associated with eth1"
    pool_id = "%[2]s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }
    network {
	network_id = "${data.xenorchestra_network.network2.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigWithThreeVIFs(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

data "xenorchestra_network" "network2" {
    name_label = "Pool-wide network associated with eth1"
    pool_id = "%[2]s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }
    network {
	network_id = "${data.xenorchestra_network.network2.id}"
    }
    network {
	network_id = "${data.xenorchestra_network.network2.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, vmName, accDefaultSr.Id)
}

func testAccVmConfigUpdateAttr(nameLabel, attr string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", nameLabel), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }

    %s
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, nameLabel, accDefaultSr.Id, attr)
}

// Terraform config that tests changes to a VM that do not require halting
// the VM prior to applying
func testAccVmConfigUpdateAttrsHaltIrrelevant(nameLabel, nameDescription, ha string, powerOn bool) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", nameLabel), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "%s"
    template = "${data.xenorchestra_template.template.id}"
    high_availability = "%s"
    auto_poweron = "%t"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, nameLabel, nameDescription, ha, powerOn, accDefaultSr.Id)
}

func testAccVmConfigUpdateAttrsHaltIrrelevantWithAffinityHost(nameLabel, nameDescription, ha string, powerOn bool) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", nameLabel), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_pool" "pool" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = data.xenorchestra_pool.pool.id
}

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "%s"
    affinity_host = "${data.xenorchestra_pool.pool.master}"
    template = "${data.xenorchestra_template.template.id}"
    high_availability = "%s"
    auto_poweron = "%t"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accTestPool.NameLabel, accDefaultNetwork.NameLabel, nameLabel, nameDescription, ha, powerOn, accDefaultSr.Id)
}

func testAccVmConfigUpdateAttrsVariableCPUAndMemory(cpus, memory int, nameLabel, nameDescription, ha string, powerOn bool) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", nameLabel), "template") + testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_pool" "pool" {
    name_label = "%s"
}

data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = data.xenorchestra_pool.pool.id
}

resource "xenorchestra_vm" "bar" {
    memory_max = %d
    cpus  = %d
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "%s"
    affinity_host = "${data.xenorchestra_pool.pool.master}"
    template = "${data.xenorchestra_template.template.id}"
    high_availability = "%s"
    auto_poweron = "%t"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, accTestPool.NameLabel, accDefaultNetwork.NameLabel, memory, cpus, nameLabel, nameDescription, ha, powerOn, accDefaultSr.Id)
}

func providerCredentials(username, password string) string {
	return fmt.Sprintf(`
provider "xenorchestra" {
  username = "%s"
  password = "%s"
}
`, username, password)
}

func testAccVmManagedResourceSetConfig(vmName string) string {
	return testAccVmManagedResourceSetWithDescriptionConfig(vmName, "")
}

func testAccVmManagedResourceSetWithDescriptionConfig(vmName, nameDescription string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccVmResourceSet(vmName) + fmt.Sprintf(`

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "%s"
    template = "${data.xenorchestra_template.template.id}"
    resource_set = "${xenorchestra_resource_set.rs.id}"
    network {
	network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, vmName, nameDescription, accDefaultSr.Id)
}

func testAccVmResourceSet(vmName string) string {
	return testAccTemplateConfig() + fmt.Sprintf(`
data "xenorchestra_network" "network" {
    name_label = "%s"
    pool_id = "%s"
}

resource "xenorchestra_resource_set" "rs" {
    name = "%s-%s"
    // This adds a non admin user to the resource set
    subjects = [
	"%s",
    ]

    // Add the template, storage repository and network to the resource set
    objects = [
	"${data.xenorchestra_template.template.id}",
	"%s",
	"${data.xenorchestra_network.network.id}",
    ]

    limit {
      type = "cpus"
      quantity = 20
    }

    limit {
      type = "disk"
      quantity = 107374182400
    }

    limit {
      type = "memory"
      quantity = 12884901888
    }
}
`, accDefaultNetwork.NameLabel, accTestPool.Id, accTestPrefix, vmName, accUser.Id, accDefaultSr.Id)
}

func testAccVmConfigWithoutResourceSet(vmName string) string {
	return testAccCloudConfigConfig(fmt.Sprintf("vm-template-%s", vmName), "template") + testAccVmResourceSet(vmName) + fmt.Sprintf(`

resource "xenorchestra_vm" "bar" {
    memory_max = 4295000000
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "%s"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
       network_id = "${data.xenorchestra_network.network.id}"
    }

    disk {
      sr_id = "%s"
      name_label = "disk 1"
      size = 10001317888
    }
}
`, vmName, accDefaultSr.Id)
}

func testAccTemplateConfig() string {
	return fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
    pool_id = "%s"
}
`, testTemplate.NameLabel, accTestPool.Id)
}

func testAccNonDefaultTemplateConfig(templateName string) string {
	return fmt.Sprintf(`
data "xenorchestra_template" "template" {
    name_label = "%s"
    pool_id = "%s"
}
`, templateName, accTestPool.Id)
}
