package client

import (
	"testing"
)

func TestGetPIFByDevice(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}

	device := "eth0"
	vlan_id := -1
	pifs, err := c.GetPIFByDevice(device, vlan_id)

	if err != nil {
		t.Fatalf("failed to find PIF with device: %s with error: %v", device, err)
	}

	pif := pifs[0]

	if pif.Device != device {
		t.Errorf("PIF's device %s should have matched %s", pif.Device, device)
	}

	if pif.Vlan != vlan_id {
		t.Errorf("PIF's vlan %d should have matched %d", pif.Vlan, vlan_id)
	}

	id := pif.Id
	pifs, err = c.GetPIF(PIF{Id: id})
	if err != nil {
		t.Fatalf("failed to find PIF with id: %s with error: %v", id, err)
	}

	if len(pifs) != 1 {
		t.Errorf("expected to find single PIF instead found: %d, %v", len(pifs), pifs)
	}
}
