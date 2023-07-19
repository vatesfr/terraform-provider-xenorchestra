package xoa

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testObjectIndex int = 1
var accTestPrefix string = "terraform-acc"
var accTestPool client.Pool
var accTestHost client.Host
var accDefaultSr client.StorageRepository
var accIsoSr client.StorageRepository
var accDefaultNetwork client.Network
var accUser client.User = client.User{Email: fmt.Sprintf("%s-%s", accTestPrefix, "regular-user")}
var testTemplate client.Template
var accTestPIF client.PIF
var disklessTestTemplate client.Template
var testIsoName string

func TestMain(m *testing.M) {
	// This leverages the existing flag defined in the terraform-plugin-sdk
	// repo defined below
	// https://github.com/hashicorp/terraform-plugin-sdk/blob/2c03a32a9d1be63a12eb18aaf12d2c5270c42346/helper/resource/testing.go#L58
	flag.Parse()
	flagSweep := flag.Lookup("sweep")

	if flagSweep != nil && flagSweep.Value.String() != "" {
		resource.TestMain(m)
	} else {
		_, runSetup := os.LookupEnv("TF_ACC")

		if runSetup {
			client.FindPoolForTests(&accTestPool)
			client.FindPIFForTests(&accTestPIF)
			client.FindTemplateForTests(&testTemplate, accTestPool.Id, "XOA_TEMPLATE")
			client.FindTemplateForTests(&disklessTestTemplate, accTestPool.Id, "XOA_DISKLESS_TEMPLATE")
			client.FindHostForTests(accTestPool.Master, &accTestHost)
			client.FindNetworkForTests(accTestPool.Id, &accDefaultNetwork)
			client.FindStorageRepositoryForTests(accTestPool, &accDefaultSr, accTestPrefix)
			client.FindIsoStorageRepositoryForTests(accTestPool, &accIsoSr, accTestPrefix, "XOA_ISO_SR")
			client.CreateUser(&accUser)
			testIsoName = os.Getenv("XOA_ISO")
		}

		code := m.Run()

		if runSetup {
			client.RemoveNetworksWithNamePrefix(accTestPrefix)("")
			client.RemoveVDIsWithPrefix(accTestPrefix)("")
			client.RemoveResourceSetsWithNamePrefix(accTestPrefix)("")
			client.RemoveTagFromAllObjects(accTestPrefix)("")
			client.RemoveUsersWithPrefix(accTestPrefix)("")
			client.RemoveCloudConfigsWithPrefix(accTestPrefix)("")
		}

		os.Exit(code)
	}
}
