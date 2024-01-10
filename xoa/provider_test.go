package xoa

import (
	"os"
	"terraform-provider-macaddress/macaddress"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccFailToStartAndHaltProviders map[string]*schema.Provider
var testAccFailToDeleteVmProviders map[string]*schema.Provider

var testAccProvider *schema.Provider
var testAccFailToStartHaltVmProvider *schema.Provider
var testAccFailToDeleteVmProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"xenorchestra": testAccProvider,
		"macaddress":   macaddress.Provider(),
	}

	testAccFailToStartHaltVmProvider = Provider()
	testAccFailToStartHaltVmProvider.ConfigureFunc = internal.GetFailToStartAndHaltXOClient
	testAccFailToStartAndHaltProviders = map[string]*schema.Provider{
		"xenorchestra": testAccFailToStartHaltVmProvider,
	}
	testAccFailToDeleteVmProvider = Provider()
	testAccFailToDeleteVmProvider.ConfigureFunc = internal.GetFailToDeleteVmXOClient
	testAccFailToDeleteVmProviders = map[string]*schema.Provider{
		"xenorchestra": testAccFailToDeleteVmProvider,
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
	if v := os.Getenv("XOA_RETRY_MAX_TIME"); v == "" {
		t.Fatal("The XOA_RETRY_MAX_TIME environment variable must be set")
	}
	if v := os.Getenv("XOA_RETRY_MODE"); v == "" {
		t.Fatal("The XOA_RETRY_MODE environment variable must be set")
	}
}
