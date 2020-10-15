package client

import "testing"

var testNetworkName string = integrationTestPrefix + "network"

func TestNetworkCompare(t *testing.T) {

	nameLabel := "network label"
	poolId := "pool id"
	cases := []struct {
		net    Network
		result bool
		other  Network
	}{
		{
			net: Network{
				NameLabel: nameLabel,
			},
			result: true,
			other: Network{
				Id:        "355ee47d-ff4c-4924-3db2-fd86ae629676",
				NameLabel: nameLabel,
				PoolId:    "355ee47d-ff4c-4924-3db2-fd86ae629676",
			},
		},
		{
			net: Network{
				NameLabel: nameLabel,
				PoolId:    poolId,
			},
			result: false,
			other: Network{
				Id:        "355ee47d-ff4c-4924-3db2-fd86ae629676",
				NameLabel: "name_label",
				PoolId:    "355ee47d-ff4c-4924-3db2-fd86ae629676",
			},
		},
		{
			net: Network{
				NameLabel: nameLabel,
				PoolId:    poolId,
			},
			result: true,
			other: Network{
				Id:        "355ee47d-ff4c-4924-3db2-fd86ae629676",
				NameLabel: nameLabel,
				PoolId:    poolId,
			},
		},
	}

	for _, test := range cases {
		net := test.net
		other := test.other
		result := test.result

		if net.Compare(other) != result {
			t.Errorf("expected network `%+v` to Compare '%t' with other `%v`", net, result, other)
		}
	}
}

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

	if net.Bridge == "" {
		t.Errorf("expected network bridge to not be an empty string")
	}

	if net.PoolId == "" {
		t.Errorf("expected network pool id to not be an empty string")
	}
}
