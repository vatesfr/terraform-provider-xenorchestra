package client

import "testing"

func TestPoolCompare(t *testing.T) {
	tests := []struct {
		other  Pool
		pool   Pool
		result bool
	}{
		{
			other: Pool{
				Id:        "sample pool id",
				NameLabel: "xenserver-ddelnano",
			},
			pool:   Pool{NameLabel: "xenserver-ddelnano"},
			result: true,
		},
		{
			other: Pool{
				Id:        "sample pool id",
				NameLabel: "does not match",
			},
			pool:   Pool{NameLabel: "xenserver-ddelnano"},
			result: false,
		},
	}

	for _, test := range tests {
		pool := test.pool
		other := test.other
		result := test.result
		if pool.Compare(other) != result {
			t.Errorf("Expected Pool %v to Compare %t to %v", pool, result, other)
		}
	}
}

func TestGetPoolByName(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}

	nameLabel := accTestPool.NameLabel
	pools, err := c.GetPoolByName(nameLabel)

	pool := pools[0]

	if err != nil {
		t.Errorf("failed to get pool with error: %v", err)
	}

	if pool.NameLabel != nameLabel {
		t.Errorf("expected pool to have name `%s` received `%s` instead.", nameLabel, pool.NameLabel)
	}

	if pool.Cpus.Cores == 0 {
		t.Errorf("expected pool cpu cores to be set")
	}

	if pool.Cpus.Sockets == 0 {
		t.Errorf("expected pool cpu sockets to be set")
	}
}
