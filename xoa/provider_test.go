package xoa

import (
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccFailToStartAndHaltProviders map[string]*schema.Provider

var testAccProvider *schema.Provider
var testAccFailToStartHaltVmProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"xenorchestra": testAccProvider,
	}

	testAccFailToStartHaltVmProvider = Provider()
	testAccFailToStartHaltVmProvider.ConfigureFunc = internal.GetFailToStartAndHaltXOClient
	testAccFailToStartAndHaltProviders = map[string]*schema.Provider{
		"xenorchestra": testAccFailToStartHaltVmProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("XOA_URL"); v == "" {
		t.Fatal("The XOA_URL environment variable must be set")
	}
	if v := os.Getenv("XOA_USER"); v == "" {
		t.Fatal("The XOA_USER environment variable must be set")
	}
	if v := os.Getenv("XOA_PASSWORD"); v == "" {
		t.Fatal("The XOA_PASSWORD environment variable must be set")
	}
	if v := os.Getenv("XOA_POOL"); v == "" {
		t.Fatal("The XOA_POOL environment variable must be set")
	}
	if v := os.Getenv("XOA_TEMPLATE"); v == "" {
		t.Fatal("The XOA_TEMPLATE environment variable must be set")
	}
	if v := os.Getenv("XOA_DISKLESS_TEMPLATE"); v == "" {
		t.Fatal("The XOA_DISKLESS_TEMPLATE environment variable must be set")
	}
	if v := os.Getenv("XOA_ISO"); v == "" {
		t.Fatal("The XOA_ISO environment variable must be set")
	}
	if v := os.Getenv("XOA_ISO_SR"); v == "" {
		t.Fatal("The XOA_ISO_SR environment variable must be set")
	}
}
