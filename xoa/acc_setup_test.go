package xoa

import (
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

var testObjectIndex int = 1
var accTestPrefix string = "terraform-acc"
var accTestPool client.Pool
var accDefaultSr client.StorageRepository

func TestMain(m *testing.M) {
	client.FindPoolForTests(&accTestPool)
	client.FindStorageRepositoryForTests(accTestPool, &accDefaultSr, accTestPrefix)
	code := m.Run()

	client.RemoveNetworksWithNamePrefix(accTestPrefix)("")
	client.RemoveResourceSetsWithNamePrefix(accTestPrefix)("")
	client.RemoveTagFromAllObjects(accTestPrefix)("")

	os.Exit(code)
}
