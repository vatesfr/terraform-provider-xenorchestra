package client

import (
	"fmt"
	"testing"
)

func TestGetVIFs(t *testing.T) {

	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}

	vmName := "XOA"
	vm, err := c.GetVm(Vm{NameLabel: vmName})

	if err != nil {
		t.Errorf("failed to get VM with error: %v", err)
	}

	vifs, err := c.GetVIFs(vm)

	for _, vif := range vifs {
		if vif.Device == "" {
			t.Errorf("expecting `Device` field to be set on VIF")
		}

		if vif.MacAddress == "" {
			t.Errorf("expecting `MacAddress` field to be set on VIF")
		}

		if vif.Network == "" {
			t.Errorf("expecting `Network` field to be set on VIF")
		}

		if vif.VmId != vm.Id {
			t.Errorf("VIF's VmId `%s` should have matched: %v", vif.VmId, vm)
		}
	}
}

func TestCreateVIF_DeleteVIF(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}

	vmName := "XOA"
	vm, err := c.GetVm(Vm{NameLabel: vmName})

	if err != nil {
		t.Errorf("failed to get VM with error: %v", err)
	}

	pif, err := c.GetPIFByDevice("eth1", -1)

	if err != nil {
		t.Errorf("failed to get PIF with error: %v", err)
	}

	vif, err := c.CreateVIF(vm, &VIF{Network: pif.Network})

	if err != nil {
		t.Errorf("failed to create VIF with error: %v", err)
	}

	fmt.Printf("VIF is %+v", vif)

	err = c.DeleteVIF(vif)

	if err != nil {
		t.Errorf("failed to delete the VIF with error: %v", err)
	}
}
