package client

import (
	"os"
	"testing"
)

func CreateResourceSet(rs ResourceSet) error {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		return err
	}
	_, err = c.CreateResourceSet(rs)
	return err
}

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
	CreateResourceSet(testResourceSet)

	code := m.Run()

	RemoveResourceSetsWithNamePrefix("xenorchestra-client")("")
	RemoveNetworksWithNamePrefix(integrationTestPrefix)("")

	os.Exit(code)
}
