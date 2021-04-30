package xoa

import (
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccFailToStartAndHaltProviders map[string]terraform.ResourceProvider

var testAccProvider *schema.Provider
var testAccFailToStartHaltVmProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"xenorchestra": testAccProvider,
	}

	testAccFailToStartHaltVmProvider = Provider().(*schema.Provider)
	testAccFailToStartHaltVmProvider.ConfigureFunc = internal.GetFailToStartAndHaltXOClient
	testAccFailToStartAndHaltProviders = map[string]terraform.ResourceProvider{
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
}
