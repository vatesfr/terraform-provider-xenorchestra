package client

import "testing"

func TestHostCompare(t *testing.T) {
	tests := []struct {
		other  Host
		host   Host
		result bool
	}{
		{
			other: Host{
				Id:        "788e1dce-44f6-4db7-ae62-185c69fecd3b",
				NameLabel: "xcp-host1-k8s.domain.eu",
			},
			host:   Host{NameLabel: "xcp-host1-k8s.domain.eu"},
			result: true,
		},
		{
			other: Host{
				Id:        "788e1dce-44f6-4db7-ae62-185c69fecd3b",
				NameLabel: "xcp-host2-k8s.domain.eu",
			},
			host:   Host{NameLabel: "xcp-host1-k8s.domain.eu"},
			result: false,
		},
	}

	for _, test := range tests {
		host := test.host
		other := test.other
		result := test.result
		if host.Compare(other) != result {
			t.Errorf("Expected Host %v to Compare %t to %v", host, result, other)
		}
	}
}

func TestGetHostByName(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	nameLabel := accTestHost.NameLabel
	hosts, err := c.GetHostByName(nameLabel)
	if err != nil {
		t.Fatalf("failed to get host with error: %v", err)
	}

	host := hosts[0]
	if host.NameLabel != nameLabel {
		t.Errorf("expected host to have name `%s` received `%s` instead.", nameLabel, host.NameLabel)
	}

}

func TestGetHostsByPoolName(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	poolName := accTestHost.Pool
	hosts, err := c.GetHostsByPoolName(poolName)
	if err != nil {
		t.Fatalf("failed to get host with error: %v", err)
	}
	if len(hosts) == 0 {
		t.Errorf("failed to find any host for pool `%s`.", poolName)
	}
	for _, host := range hosts {
		if host["pool"] != poolName {
			t.Errorf("expected pool to have name `%s` received `%s` instead.", poolName, hosts[0]["pool"])
		}

	}
}
