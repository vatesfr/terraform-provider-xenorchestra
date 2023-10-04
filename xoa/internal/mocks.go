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
	return newFailToStartAndHaltClient(client.GetConfigFromEnv())
}
