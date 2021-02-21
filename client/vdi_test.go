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

	disks, err := c.GetDisks(&accVm)

	if err != nil {
		t.Fatalf("failed to get disks of VM with error: %v", err)
	}

	if len(disks) <= 0 {
		t.Fatalf("failed to find disks for Vm")
	}

	if !validateDisk(disks[0]) {
		t.Errorf("failed to validate that disks contained expected data")
	}
}

func validateDisk(disk Disk) bool {
	if disk.Id == "" {
		return false
	}

	if disk.VBD.PoolId == "" {
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

	diskNameLabel := fmt.Sprintf("%stesting", integrationTestPrefix)
	diskId, err := c.CreateDisk(
		accVm,
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

	disks, err := c.GetDisks(&accVm)

	for _, disk := range disks {
		if disk.NameLabel != diskNameLabel {
			continue
		}

		err = c.DeleteDisk(accVm, disk)

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

	disks, err := c.GetDisks(&accVm)

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
