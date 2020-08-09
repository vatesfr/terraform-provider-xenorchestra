package client

import "testing"

func TestPoolCompare(t *testing.T) {
	tests := []struct {
		object map[string]interface{}
		pool   Pool
		result bool
	}{
		{
			object: map[string]interface{}{
				"name_label": "xenserver-ddelnano",
				"$poolId":    "Sample pool id",
			},
			pool:   Pool{NameLabel: "xenserver-ddelnano"},
			result: true,
		},
		{
			object: map[string]interface{}{
				"name_label": "Does not match",
				"$poolId":    "Sample pool id",
			},
			pool:   Pool{NameLabel: "xenserver-ddelnano"},
			result: false,
		},
	}

	for _, test := range tests {
		pool := test.pool
		object := test.object
		result := test.result
		if pool.Compare(object) != result {
			t.Errorf("Expected Pool %v to Compare %t to %v", pool, result, object)
		}
	}
}

func TestGetPoolByName(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}

	nameLabel := "xenserver-ddelnano"

	pool, err := c.GetPoolByName(nameLabel)

	if err != nil {
		t.Errorf("failed to get pool with error: %v", err)
	}

	if pool.NameLabel != nameLabel {
		t.Errorf("expected pool to have name `%s` received `%s` instead.", nameLabel, pool.NameLabel)
	}
}
