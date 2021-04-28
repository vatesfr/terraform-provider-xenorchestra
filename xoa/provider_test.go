package xoa

import (
	"errors"
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProviders2 map[string]terraform.ResourceProvider

var testAccProvider *schema.Provider
var testAccProvider2 *schema.Provider

type testClient struct {
	*client.Client
}

func (c testClient) HaltVm(vmReq client.Vm) error {
	return errors.New("This method shouldn't be called")
}
func (c testClient) StartVm(id string) error {
	return errors.New("This method shouldn't be called")
}

func newTestClient(config client.Config) (client.XOClient, error) {
	xoClient, err := client.NewClient(config)

	if err != nil {
		return nil, err
	}

	c := xoClient.(*client.Client)

	return &testClient{c}, nil
}

func configureFn(d *schema.ResourceData) (interface{}, error) {
	url := d.Get("url").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	insecure := d.Get("insecure").(bool)
	config := client.Config{
		Url:                url,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecure,
	}
	return newTestClient(config)
}

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"xenorchestra": testAccProvider,
	}

	testAccProvider2 = Provider().(*schema.Provider)
	testAccProvider2.ConfigureFunc = configureFn
	testAccProviders2 = map[string]terraform.ResourceProvider{
		"xenorchestra": testAccProvider2,
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
