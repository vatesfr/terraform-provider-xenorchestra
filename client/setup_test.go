package client

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	createResourceSet()
	code := m.Run()

	tearDownResourceSet()

	os.Exit(code)
}
