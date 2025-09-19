package xoa

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

var testObjectIndex int = 1
var accTestPrefix string = "terraform-acc"
var accTestXoToken string = os.Getenv("BYPASS_XOA_TOKEN")
var accTestPool client.Pool
var accTestHost client.Host
var accDefaultSr client.StorageRepository
var accIsoSr client.StorageRepository
var accDefaultNetwork client.Network
var accUser client.User = client.User{Email: fmt.Sprintf("%s-%s", accTestPrefix, "regular-user")}
var testTemplate client.Template
var accTestPIF client.PIF
var disklessTestTemplate client.Template
var testIso client.VDI

func TestMain(m *testing.M) {
	// This leverages the existing flag defined in the terraform-plugin-sdk
	// repo defined below
	// https://github.com/hashicorp/terraform-plugin-sdk/blob/2c03a32a9d1be63a12eb18aaf12d2c5270c42346/helper/resource/testing.go#L58
	flag.Parse()
	flagSweep := flag.Lookup("sweep")

	if flagSweep != nil && flagSweep.Value.String() != "" {
		_, runSetup := os.LookupEnv("TF_ACC")

		if runSetup {
			fmt.Println("Running sweeping")
			resource.TestMain(m)

			client.FindPoolForTests(&accTestPool)
			client.FindPIFForTests(&accTestPIF)
			client.FindTemplateForTests(&testTemplate, accTestPool.Id, "XOA_TEMPLATE")
			client.FindTemplateForTests(&disklessTestTemplate, accTestPool.Id, "XOA_DISKLESS_TEMPLATE")
			client.FindHostForTests(accTestPool.Master, &accTestHost)
			client.FindNetworkForTests(accTestPool.Id, &accDefaultNetwork)
			client.FindStorageRepositoryForTests(accTestPool, &accDefaultSr, accTestPrefix)
			client.FindIsoStorageRepositoryForTests(accTestPool, &accIsoSr, accTestPrefix, "XOA_ISO_SR")
			client.CreateUserForTests(&accUser)
			client.FindVDIForTests(accTestPool, &testIso, "XOA_ISO")
			fmt.Printf("Found the following pool: %v sr: %v\n", accTestPool, accDefaultSr)

			code := m.Run()
			client.RemoveNetworksWithNamePrefixForTests(accTestPrefix)("")
			client.RemoveVDIsWithPrefixForTests(accTestPrefix)("")
			client.RemoveResourceSetsWithNamePrefixForTests(accTestPrefix)("")
			client.RemoveTagFromAllObjectsForTests(accTestPrefix)("")
			client.RemoveUsersWithPrefixForTests(accTestPrefix)("")
			client.RemoveCloudConfigsWithPrefix(accTestPrefix)("")
			os.Exit(code)
		}
	}

}
