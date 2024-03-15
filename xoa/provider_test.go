package xoa

import (
	"os"
	"testing"

	"github.com/vatesfr/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccTokenAuthProviders map[string]*schema.Provider
var testAccFailToStartAndHaltProviders map[string]*schema.Provider
var testAccFailToDeleteVmProviders map[string]*schema.Provider

var testAccProvider *schema.Provider
var testAccTokenAuthProvider *schema.Provider
var testAccFailToStartHaltVmProvider *schema.Provider
var testAccFailToDeleteVmProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"xenorchestra": testAccProvider,
	}

	testAccTokenAuthProvider = createTokenAuthProvider()
	testAccTokenAuthProviders = map[string]*schema.Provider{
		"xenorchestra": testAccTokenAuthProvider,
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

func createTokenAuthProvider() *schema.Provider {
	provider := Provider()

	// The test suite runs in an environment where the XOA_USER and XOA_PASSWORD environment
	// variables are set. Therefore the DefaultFunc's and ConflictsWith's will think that
	// username, password and token were supplied and will fail validation. The patching
	// below allows this test provider to think only token auth is supplied (ConflictsWith changes)
	// and prevents the username and password from being passed through (DefaultFunc changes).
	var f schema.SchemaDefaultFunc = func() (interface{}, error) { return "", nil }
	provider.Schema["username"].DefaultFunc = f
	provider.Schema["username"].ConflictsWith = []string{}

	provider.Schema["password"].DefaultFunc = f
	provider.Schema["password"].ConflictsWith = []string{}

	provider.Schema["token"].ConflictsWith = []string{}
	provider.Schema["token"].DefaultFunc = schema.EnvDefaultFunc("BYPASS_XOA_TOKEN", nil)
	return provider
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("XOA_URL"); v == "" {
		t.Fatal("The XOA_URL environment variable must be set")
	}

	user := os.Getenv("XOA_USER")
	password := os.Getenv("XOA_PASSWORD")
	token := os.Getenv("BYPASS_XOA_TOKEN")

	if token == "" && (user == "" || password == "") {
		t.Fatal("One of the following environment variable(s) must be set: XOA_USER and XOA_PASSWORD or BYPASS_XOA_TOKEN")
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
