package xoa

import (
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

var testObjectIndex int = 1
var accTestPrefix string = "terraform-acc"
var accTestPool client.Pool
var accTestHost client.Host
var accDefaultSr client.StorageRepository
var accDefaultNetwork client.Network
var testTemplate client.Template
var disklessTestTemplate client.Template
var testIsoName string

func TestMain(m *testing.M) {
	_, runSetup := os.LookupEnv("TF_ACC")

	if runSetup {
		client.FindPoolForTests(&accTestPool)
		client.FindTemplateForTests(&testTemplate, accTestPool.Id, "XOA_TEMPLATE")
		client.FindTemplateForTests(&disklessTestTemplate, accTestPool.Id, "XOA_DISKLESS_TEMPLATE")
		client.FindHostForTests(accTestPool.Master, &accTestHost)
		client.FindNetworkForTests(accTestPool.Id, &accDefaultNetwork)
		client.FindStorageRepositoryForTests(accTestPool, &accDefaultSr, accTestPrefix)
		testIsoName = os.Getenv("XOA_ISO")
	}

	code := m.Run()

	if runSetup {
		client.RemoveNetworksWithNamePrefix(accTestPrefix)("")
		client.RemoveResourceSetsWithNamePrefix(accTestPrefix)("")
		client.RemoveTagFromAllObjects(accTestPrefix)("")
		client.RemoveUsersWithPrefix(accTestPrefix)("")
		client.RemoveCloudConfigsWithPrefix(accTestPrefix)("")
	}

	os.Exit(code)
}
