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
var testTemplate client.Template

func TestMain(m *testing.M) {
	client.FindTemplateForTests(&testTemplate)
	client.FindPoolForTests(&accTestPool)
	client.FindHostForTests(&accTestHost)
	client.FindStorageRepositoryForTests(accTestPool, &accDefaultSr, accTestPrefix)
	code := m.Run()

	client.RemoveNetworksWithNamePrefix(accTestPrefix)("")
	client.RemoveResourceSetsWithNamePrefix(accTestPrefix)("")
	client.RemoveTagFromAllObjects(accTestPrefix)("")
	client.RemoveUsersWithPrefix(accTestPrefix)("")
	client.RemoveCloudConfigsWithPrefix(accTestPrefix)("")

	os.Exit(code)
}
