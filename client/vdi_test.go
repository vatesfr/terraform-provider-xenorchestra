package client

import (
	"fmt"
	"testing"
)

func TestGetVmDisks(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	// TODO: Create Vm for client tests / optionally allow
	// for running tests against a running Vm
	disks, err := c.GetDisks(&Vm{
		Id: "75e7a443-f9c5-0afa-29ef-c15a33690a00",
	})

	if err != nil {
		t.Fatalf("failed to get disks of VM with error: %v", err)
	}

	if len(disks) <= 0 {
		t.Fatalf("failed to find disks for Vm")
	}

	fmt.Printf("Found disks with %+v", disks)
	if !validateDisk(disks[0]) {
		t.Errorf("failed to validate that disks contained expected data")
	}
}

func validateDisk(disk Disk) bool {
	if disk.Id == "" {
		return false
	}

	if disk.PoolId == "" {
		return false
	}

	if disk.Device == "" {
		return false
	}

	if disk.NameLabel == "" {
		return false
	}
	return true
}

func TestCreateDiskAndDeleteDisk(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	// TODO: Create Vm for client tests / optionally allow
	// for running tests against a running Vm
	vm := Vm{
		Id: "75e7a443-f9c5-0afa-29ef-c15a33690a00",
	}
	diskNameLabel := fmt.Sprintf("%stesting", integrationTestPrefix)
	diskId, err := c.CreateDisk(
		vm,
		Disk{
			VBD{},
			VDI{
				NameLabel: diskNameLabel,
				Size:      10000,
				SrId:      accDefaultSr.Id,
			},
		},
	)

	if err != nil {
		t.Fatalf("failed to create disk with error: %v", err)
	}

	disks, err := c.GetDisks(&vm)

	for _, disk := range disks {
		if disk.NameLabel != diskNameLabel {
			continue
		}

		err = c.DeleteDisk(vm, disk)

		if err != nil {
			t.Errorf("failed to delete disk with id: %s with error: %v", diskId, err)
		}
	}
}

func TestDisconnectDiskAndConnectDisk(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	// TODO: Create Vm for client tests / optionally allow
	// for running tests against a running Vm
	vm := Vm{
		Id: "75e7a443-f9c5-0afa-29ef-c15a33690a00",
	}
	disks, err := c.GetDisks(&vm)

	if err != nil {
		t.Fatalf("failed to retrieve disks with error: %v", err)
	}

	if err := c.DisconnectDisk(disks[1]); err != nil {
		t.Fatalf("failed to disconnect disk: %+v with error: %v", disks[1], err)
	}

	if err := c.ConnectDisk(disks[1]); err != nil {
		t.Errorf("failed to connect disk: %+v with error: %v", disks[1], err)
	}
}
