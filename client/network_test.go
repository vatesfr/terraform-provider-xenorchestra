package client

import "testing"

var testNetworkName string = integrationTestPrefix + "network"

func TestNetworkCompare(t *testing.T) {

	nameLabel := "network label"
	poolId := "pool id"
	cases := []struct {
		net    Network
		result bool
		obj    map[string]interface{}
	}{
		{
			net: Network{
				NameLabel: nameLabel,
			},
			result: true,
			obj: map[string]interface{}{
				"id":         "355ee47d-ff4c-4924-3db2-fd86ae629676",
				"name_label": nameLabel,
				"$poolId":    "355ee47d-ff4c-4924-3db2-fd86ae629676",
			},
		},
		{
			net: Network{
				NameLabel: nameLabel,
				PoolId:    poolId,
			},
			result: false,
			obj: map[string]interface{}{
				"id":         "355ee47d-ff4c-4924-3db2-fd86ae629676",
				"name_label": nameLabel,
				"$poolId":    "355ee47d-ff4c-4924-3db2-fd86ae629676",
			},
		},
		{
			net: Network{
				NameLabel: nameLabel,
				PoolId:    poolId,
			},
			result: true,
			obj: map[string]interface{}{
				"id":         "355ee47d-ff4c-4924-3db2-fd86ae629676",
				"name_label": nameLabel,
				"$poolId":    poolId,
			},
		},
	}

	for _, test := range cases {
		net := test.net
		obj := test.obj
		result := test.result

		if net.Compare(obj) != result {
			t.Errorf("expected network `%+v` to Compare '%t' with object `%v`", net, result, obj)
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
}
