package client

import (
	"fmt"
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
	findPoolForTests()
	CreateNetwork()
	code := m.Run()

	RemoveNetworksWithNamePrefix(integrationTestPrefix)("")

	os.Exit(code)
}

func findPoolForTests() {
	poolName, found := os.LookupEnv("XOA_POOL")

	if !found {
		fmt.Println("The XOA_POOL environment variable must be set")
		os.Exit(-1)
	}
	c, _ := NewClient(GetConfigFromEnv())
	var err error
	accTestPool, err = c.GetPoolByName(poolName)

	if err != nil {
		fmt.Printf("failed to find a pool with name: %v with error: %v\n", poolName, err)
		os.Exit(-1)
	}
}
