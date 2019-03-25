package client

import "testing"

func TestGetPIFByDevice(t *testing.T) {
	c, err := NewClient()

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}

	device := "eth0"
	pif, err := c.GetPIFByDevice(device)

	if err != nil {
		t.Errorf("failed to find PIF with device: %s with error: %v", device, err)
	}

	if pif.Device != device {
		t.Errorf("PIF's device %s should have matched %s", pif.Device, device)
	}
}
