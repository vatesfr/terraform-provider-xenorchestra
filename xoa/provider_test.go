package xoa

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"xenorchestra": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("XOA_HOST"); v == "" {
		t.Fatal("The XOA_HOST environment variable must be set")
	}
	if v := os.Getenv("XOA_USER"); v == "" {
		t.Fatal("The XOA_USER environment variable must be set")
	}
	if v := os.Getenv("XOA_PASSWORD"); v == "" {
		t.Fatal("The XOA_PASSWORD environment variable must be set")
	}
}
