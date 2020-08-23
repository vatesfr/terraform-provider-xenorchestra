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

	fmt.Printf("%+v", vifs)
}
