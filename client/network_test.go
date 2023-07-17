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

func TestCreateNetwork_DeleteNetwork(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	var vlan = 100
	var netReq Network = Network{
		NameLabel:   integrationTestPrefix + "created-network",
		PifIds:      []string{accTestPif.Id},
		Description: "Network created by integration tests",
		Mtu:         1500,
		PoolId:      accTestPool.Id,
	}

	resultNet, err := c.CreateNetwork(netReq, vlan)

	if err != nil {
		t.Fatalf("failed to create network with error: %v", err)
	}

	if resultNet.Id == "" {
		t.Errorf("expected network Id to not be empty")
	}

	if resultNet.NameLabel != netReq.NameLabel {
		t.Errorf("expected network name_label `%s` to match `%s`", resultNet.NameLabel, netReq.NameLabel)
	}

	if len(resultNet.PifIds) == 0 {
		t.Errorf("expected network pifs to not be empty")
	}

	if resultNet.Description != netReq.Description {
		t.Errorf("expected network description `%s` to match `%s`", resultNet.Description, netReq.Description)
	}

	if resultNet.Mtu != netReq.Mtu {
		t.Errorf("expected network mtu `%d` to match `%d`", resultNet.Mtu, netReq.Mtu)
	}

	if resultNet.PoolId != netReq.PoolId {
		t.Errorf("expected network poolId `%s` to match `%s`", resultNet.PoolId, netReq.PoolId)
	}

	//Creating a Network while specifying a PIF ID and a VLAN creates clone of the given PIF with the given VLAN
	var pifs []PIF
	pifs, err = c.GetPIF(PIF{Host: accTestHost.Id, Device: "eth0", Vlan: 100, Id: resultNet.PifIds[0]})

	if err != nil {
		t.Fatalf("failed to get pif of with id %s with error: %v", resultNet.PifIds[0], err)
	}

	if len(pifs) == 0 || pifs[0].Vlan != vlan {
		t.Errorf("expected network VLAN `%d` to match `%d`", pifs[0].Vlan, vlan)
	}

	err = c.DeleteNetwork(resultNet.Id)

	if err != nil {
		t.Errorf("failed to delete the network with error: %v", err)
	}
}
