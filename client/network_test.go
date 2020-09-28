package client

import "testing"

var testNetworkName string = integrationTestPrefix + "network"

func TestGetNetwork(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	net, err := c.GetNetwork(Network{
		NameLabel: testNetworkName,
	})

	if err != nil {
		t.Fatalf("failed to retrieve network `%s` with error: %v", testNetworkName, err)
	}

	if net == nil {
		t.Fatalf("should have received network, instead received nil")
	}

	if net.NameLabel != testNetworkName {
		t.Errorf("expected network name_label `%s` to match `%s`", net.NameLabel, testNetworkName)
	}
}
