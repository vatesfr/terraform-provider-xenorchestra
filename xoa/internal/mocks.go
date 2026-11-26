package internal

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

// failToStartAndHaltVmXOClient is a mock client used to ensure that HaltVm
// and StartVm is not called. This is useful for tests that need to ensure that
// a Vm is modified without rebooting for CPU or memory changes
type failToStartAndHaltVmXOClient struct {
	*client.Client
}

func (c failToStartAndHaltVmXOClient) HaltVm(id string) error {
	return errors.New("This method shouldn't be called")
}
func (c failToStartAndHaltVmXOClient) StartVm(id string) error {
	return errors.New("This method shouldn't be called")
}

func newFailToStartAndHaltClient(config client.Config) (client.XOClient, error) {
	xoClient, err := client.NewClient(config)

	if err != nil {
		return nil, err
	}

	c := xoClient.(*client.Client)

	return &failToStartAndHaltVmXOClient{c}, nil
}

func GetFailToStartAndHaltXOClient(d *schema.ResourceData) (interface{}, error) {
	return newFailToStartAndHaltClient(client.GetConfigFromEnv())
}

type failToDeleteVmXOClient struct {
	*client.Client
}

func (c *failToDeleteVmXOClient) DeleteVm(id string) error {
	return errors.New("client.DeleteVm should not be called")
}

func newFailToDeleteVmClient(config client.Config) (client.XOClient, error) {
	xoClient, err := client.NewClient(config)

	if err != nil {
		return nil, err
	}

	c := xoClient.(*client.Client)

	return &failToDeleteVmXOClient{c}, nil
}

func GetFailToDeleteVmXOClient(d *schema.ResourceData) (interface{}, error) {
	return newFailToDeleteVmClient(client.GetConfigFromEnv())
}

// shutdownBeforeDestroyXOClient is a mock client that only lets a VM be deleted
// once it is halted, and fails VM updates. It verifies that the xenorchestra_vm
// resource cleanly shuts down VMs before destroying them when
// shutdown_before_destroy is set, and that toggling that attribute does not
// go through a VM update.
type shutdownBeforeDestroyXOClient struct {
	*client.Client
}

func (c *shutdownBeforeDestroyXOClient) DeleteVm(id string) error {
	vm, err := c.GetVm(client.Vm{Id: id})
	if err != nil {
		return err
	}

	if vm.PowerState != client.HaltedPowerState {
		return fmt.Errorf("mock client received a VM in power state %q instead of %q: clean shutdown was bypassed", vm.PowerState, client.HaltedPowerState)
	}

	return c.Client.DeleteVm(id)
}

func (c *shutdownBeforeDestroyXOClient) UpdateVm(vmReq client.Vm) (*client.Vm, error) {
	return nil, errors.New("client.UpdateVm should not be called")
}

func newShutdownBeforeDestroyClient(config client.Config) (client.XOClient, error) {
	xoClient, err := client.NewClient(config)

	if err != nil {
		return nil, err
	}

	c := xoClient.(*client.Client)

	return &shutdownBeforeDestroyXOClient{c}, nil
}

func GetShutdownBeforeDestroyXOClient(d *schema.ResourceData) (interface{}, error) {
	return newShutdownBeforeDestroyClient(client.GetConfigFromEnv())
}
