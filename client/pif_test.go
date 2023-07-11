package client

import (
	"fmt"
	"os"
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
}

func FindPIFForTests(hostId string, pif *PIF) {
	var pifs []PIF

	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	//Assuming
	pifs, err = c.GetPIF(PIF{Host: hostId, Device: "eth0", Vlan: -1})

	if err != nil || len(pifs) != 1 {
		fmt.Printf("failed to find a PIF on hostId: %v with device eth0 and Vlan -1 with error: %v\n", hostId, err)
		os.Exit(-1)
	}

	*pif = pifs[0]
}
