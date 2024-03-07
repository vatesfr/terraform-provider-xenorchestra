package client

import (
	"fmt"
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

func CreateNetwork(network *Network) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		fmt.Printf("[ERROR] Failed to create network with '%v'\n", err)
		os.Exit(1)
	}

	net, err := c.CreateNetwork(CreateNetworkRequest{
		Name: testNetworkName,
		Pool: accTestPool.Id,
	})

	if err != nil {
		fmt.Printf("[ERROR] Failed to create network with '%v'\n", err)
		os.Exit(1)
	}
	*network = *net
}

var integrationTestPrefix string = "xo-go-client-"
var accTestPool Pool
var accTestHost Host
var accDefaultSr StorageRepository
var accDefaultNetwork Network
var testTemplate Template
var disklessTestTemplate Template
var accVm Vm

func TestMain(m *testing.M) {
	FindPoolForTests(&accTestPool)
	FindTemplateForTests(&testTemplate, accTestPool.Id, "XOA_TEMPLATE")
	FindTemplateForTests(&disklessTestTemplate, accTestPool.Id, "XOA_DISKLESS_TEMPLATE")
	FindHostForTests(accTestPool.Master, &accTestHost)
	FindStorageRepositoryForTests(accTestPool, &accDefaultSr, integrationTestPrefix)
	CreateNetwork(&accDefaultNetwork)
	FindOrCreateVmForTests(&accVm, accTestPool.Id, accDefaultSr.Id, testTemplate.Id, integrationTestPrefix)
	CreateResourceSet(testResourceSet)

	code := m.Run()

	RemoveResourceSetsWithNamePrefix(integrationTestPrefix)("")
	RemoveNetworksWithNamePrefix(integrationTestPrefix)("")

	os.Exit(code)
}
