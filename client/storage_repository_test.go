package client

import (
	"testing"
)

func TestStorageRepositoryCompare(t *testing.T) {
	tests := []struct {
		object map[string]interface{}
		sr     StorageRepository
		result bool
	}{
		{
			object: map[string]interface{}{
				"name_label": "Test",
				"$poolId":    "Sample pool id",
			},
			sr:     StorageRepository{NameLabel: "Test"},
			result: true,
		},
		{
			object: map[string]interface{}{
				"name_label": "Test",
				"$poolId":    "Pool A",
			},
			sr:     StorageRepository{NameLabel: "Test", PoolId: "Pool A"},
			result: true,
		},
		{
			object: map[string]interface{}{
				"name_label": "Test",
				"$poolId":    "Pool A",
			},
			sr:     StorageRepository{NameLabel: "Test", PoolId: "Pool A"},
			result: true,
		},
	}

	for _, test := range tests {
		sr := test.sr
		object := test.object
		result := test.result
		if sr.Compare(object) != result {
			t.Errorf("Expected Storage Repository %v to Compare %t to %v", sr, result, object)
		}
	}
}

func TestGetStorageRepositoryByType(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}
	expectedNameLabel := "XenServer Tools"
	sr := StorageRepository{
		NameLabel: expectedNameLabel,
	}
	sr, err = c.GetStorageRepository(sr)

	if err != nil {
		t.Errorf("failed to get storage repository with error: %v", err)
	}

	if sr.NameLabel != expectedNameLabel {
		t.Errorf("expected storage repository to have name `%s` received `%s` instead.", expectedNameLabel, sr.NameLabel)
	}
}
