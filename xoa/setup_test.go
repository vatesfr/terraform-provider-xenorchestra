package xoa

import (
	"os"
	"testing"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

var testResourceSetName string = "terraform-acc-resource-set"
var testResourceSet = client.ResourceSet{
	Name: testResourceSetName,
	Limits: client.ResourceSetLimits{
		Cpus: client.ResourceSetLimit{
			Total:     1,
			Available: 2,
		},
		Disk: client.ResourceSetLimit{
			Total:     1,
			Available: 2,
		},
		Memory: client.ResourceSetLimit{
			Total:     1,
			Available: 2,
		},
	},
	Subjects: []string{},
	Objects:  []string{},
}

func CreateResourceSet(rs client.ResourceSet) error {

	c, err := client.NewClient(client.GetConfigFromEnv())

	if err != nil {
		return err
	}

	_, err = c.CreateResourceSet(rs)
	return err
}

func TearDownResourceSet(rs client.ResourceSet) error {

	c, err := client.NewClient(client.GetConfigFromEnv())

	if err != nil {
		return err
	}

	return c.DeleteResourceSet(rs)
}
func TestMain(m *testing.M) {
	CreateResourceSet(testResourceSet)
	code := m.Run()

	TearDownResourceSet(testResourceSet)

	os.Exit(code)
}
