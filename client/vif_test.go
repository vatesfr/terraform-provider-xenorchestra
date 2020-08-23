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

	vm, err := c.GetVm("d2efe162-35b3-f84f-8a59-6064b6875b61")

	if err != nil {
		t.Errorf("failed to get VM with error: %v", err)
	}

	vifs, err := c.GetVIFs(vm)

	fmt.Printf("%+v", vifs)
}
