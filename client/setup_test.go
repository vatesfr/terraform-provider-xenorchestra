package client

import (
	"os"
	"testing"
)

func CreateNetwork() error {

	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		return err
	}

	_, err = c.CreateNetwork(Network{
		NameLabel: testNetworkName,
		PoolId:    accTestPool.Id,
	})
	return err
}

var integrationTestPrefix string = "xenorchestra-client-"
var accTestPool Pool

func TestMain(m *testing.M) {
	FindPoolForTests(&accTestPool)
	CreateNetwork()
	code := m.Run()

	RemoveNetworksWithNamePrefix(integrationTestPrefix)("")

	os.Exit(code)
}
