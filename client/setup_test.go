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

func CreateNetwork(network *Network) error {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		return err
	}

	net, err := c.CreateNetwork(Network{
		NameLabel: testNetworkName,
		PoolId:    accTestPool.Id,
	})

	if err != nil {
		return err
	}
	*network = *net
	return nil
}

var integrationTestPrefix string = "xenorchestra-client-"
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
