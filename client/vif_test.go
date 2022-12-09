package client

import (
	"testing"
)

func TestGetVIFs(t *testing.T) {

	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	vifs, err := c.GetVIFs(&accVm)

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

		if vif.VmId != accVm.Id {
			t.Errorf("VIF's VmId `%s` should have matched: %v", vif.VmId, accVm)
		}

		if len(vif.Device) == 0 {
			t.Errorf("expecting `Device` field to be set on VIF instead received: %s", vif.Device)
		}

		// 		if !vif.Attached {
		// 			t.Errorf("expecting `Attached` field to be true on VIF instead received: %t", vif.Attached)
		// 		}
	}
}

func TestGetVIF(t *testing.T) {

	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	vifs, err := c.GetVIFs(&accVm)

	expectedVIF := vifs[0]

	vif, err := c.GetVIF(&VIF{
		MacAddress: expectedVIF.MacAddress,
	})

	if err != nil {
		t.Fatalf("failed to get VIF with error: %v", err)
	}

	if vif.MacAddress != expectedVIF.MacAddress {
		t.Errorf("expected VIF: %v does not match the VIF we received %v", expectedVIF, vif)
	}
}

func TestCreateVIF_DeleteVIF(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	vm, err := c.GetVm(Vm{Id: accVm.Id})

	if err != nil {
		t.Fatalf("failed to get VM with error: %v", err)
	}

	vif, err := c.CreateVIF(vm, &VIF{Network: accDefaultNetwork.Id})

	if err != nil {
		t.Fatalf("failed to create VIF with error: %v", err)
	}

	err = c.DeleteVIF(vif)

	if err != nil {
		t.Errorf("failed to delete the VIF with error: %v", err)
	}
}
