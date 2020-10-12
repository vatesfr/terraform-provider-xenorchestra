package xoa

import (
	"fmt"
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

var testObjectIndex int = 1
var accTestPrefix string = "terraform-acc-test-"
var accTestPool client.Pool
var accDefaultSr client.StorageRepository
var accTemplateName string

func TestMain(m *testing.M) {
	client.FindPoolForTests(&accTestPool)
	client.FindStorageRepositoryForTests(accTestPool, &accDefaultSr, accTestPrefix)
	var found bool
	accTemplateName, found = os.LookupEnv("XOA_TEMPLATE")
	if !found {
		fmt.Println("The XOA_TEMPLATE environment variable must be set for the tests")
		os.Exit(-1)
	}
	code := m.Run()

	client.RemoveNetworksWithNamePrefix("terraform-acc")("")
	client.RemoveTagFromAllObjects(accTestPrefix)("")
	os.Exit(code)
}
