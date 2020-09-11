package xoa

import (
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

func TestMain(m *testing.M) {
	code := m.Run()

	client.RemoveResourceSetsWithNamePrefix("terraform-acc")("")

	os.Exit(code)
}
