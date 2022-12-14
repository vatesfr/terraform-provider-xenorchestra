package internal

import (
	"errors"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	url := d.Get("url").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	insecure := d.Get("insecure").(bool)
	config := client.Config{
		Url:                url,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecure,
	}
	return newFailToStartAndHaltClient(config)
}

// gracefulVmTerminationClient is a mock client that verifies Vms are only terminated
// if they are stopped first. This is necessary to validate the xenorchestra_vm resource's
// graceful termination functionality works.
type gracefulVmTerminationClient struct {
	*client.Client
}

func (c gracefulVmTerminationClient) DeleteVm(id string) error {
	vm, err := c.GetVm(client.Vm{Id: id})

	if err != nil {
		return err
	}

	if vm.PowerState != "Halted" {
		return errors.New("mock client did not receive a stopped Vm. Graceful termination was bypassed!\n")
	}

	return c.Client.DeleteVm(id)
}

func newGracefulVmTerminationClient(config client.Config) (client.XOClient, error) {
	xoClient, err := client.NewClient(config)

	if err != nil {
		return nil, err
	}

	c := xoClient.(*client.Client)

	return &gracefulVmTerminationClient{c}, nil
}

func GetGracefulVmTerminationClient(d *schema.ResourceData) (interface{}, error) {
	url := d.Get("url").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	insecure := d.Get("insecure").(bool)
	config := client.Config{
		Url:                url,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecure,
	}
	return newGracefulVmTerminationClient(config)
}
