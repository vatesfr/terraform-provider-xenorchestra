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

func TearDownResourceSet(rs ResourceSet) error {

	c, err := NewClient(GetConfigFromEnv())

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
