package xoa

import (
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

var testObjectIndex int = 1
var accTestPrefix string = "terraform-acc-test-"
var accTestPool client.Pool

func TestMain(m *testing.M) {
	client.FindPoolForTests(&accTestPool)
	code := m.Run()

	client.RemoveNetworksWithNamePrefix("terraform-acc")("")

	os.Exit(code)
}
