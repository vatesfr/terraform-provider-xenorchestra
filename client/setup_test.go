package client

import (
	"os"
	"testing"
)

func CreateNetwork(rs Network) error {

	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		return err
	}

	_, err = c.CreateNetwork(rs)
	return err
}

var integrationTestPrefix string = "xenorchestra-client-"

func TestMain(m *testing.M) {
	CreateNetwork(testNetwork)
	code := m.Run()

	RemoveNetworksWithNamePrefix(integrationTestPrefix)("")

	os.Exit(code)
}
