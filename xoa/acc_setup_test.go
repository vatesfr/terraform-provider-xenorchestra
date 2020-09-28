package xoa

import (
	"fmt"
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

var testObjectIndex int = 1
var accTestPrefix string = "terraform-acc-test-"
var accTestPool client.Pool

func TestMain(m *testing.M) {
	findPoolForTests()
	code := m.Run()

	client.RemoveNetworksWithNamePrefix("terraform-acc")("")

	os.Exit(code)
}

func findPoolForTests() {
	poolName, found := os.LookupEnv("XOA_POOL")

	if !found {
		fmt.Println("The XOA_POOL environment variable must be set")
		os.Exit(-1)
	}
	c, _ := client.NewClient(client.GetConfigFromEnv())
	var err error
	accTestPool, err = c.GetPoolByName(poolName)

	if err != nil {
		fmt.Printf("failed to find a pool with name: %v with error: %v\n", poolName, err)
		os.Exit(-1)
	}
}
