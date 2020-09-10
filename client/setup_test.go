package client

import (
	"os"
	"testing"
)

func CreateResourceSet(rs ResourceSet) error {

	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		return err
	}

	_, err = c.CreateResourceSet(rs)
	return err
}
func TestMain(m *testing.M) {
	CreateResourceSet(testResourceSet)
	code := m.Run()

	RemoveResourceSetsWithNamePrefix("terraform-acc")("")

	os.Exit(code)
}
