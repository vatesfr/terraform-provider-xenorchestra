package client

import (
	"testing"
)

func TestStorageRepositoryCompare(t *testing.T) {
	tests := []struct {
		other  StorageRepository
		sr     StorageRepository
		result bool
	}{
		{
			other: StorageRepository{
				NameLabel: "Test",
				PoolId:    "Sample pool id",
			},
			sr:     StorageRepository{NameLabel: "Test"},
			result: true,
		},
		{
			other: StorageRepository{
				NameLabel: "Test",
				PoolId:    "Pool A",
			},
			sr:     StorageRepository{NameLabel: "Test", PoolId: "Pool A"},
			result: true,
		},
		{
			other: StorageRepository{
				NameLabel: "Test",
				PoolId:    "does not match",
			},
			sr:     StorageRepository{NameLabel: "Test", PoolId: "Pool A"},
			result: false,
		},
		{
			other: StorageRepository{
				NameLabel: "Test",
				PoolId:    "Pool A",
			},
			sr: StorageRepository{
				NameLabel: "Test",
				PoolId:    "Pool A",
				Tags:      []string{"tag1"},
			},
			result: false,
		},
	}

	for _, test := range tests {
		sr := test.sr
		other := test.other
		result := test.result
		if sr.Compare(other) != result {
			t.Errorf("Expected Storage Repository %v to Compare %t to %v", sr, result, other)
		}
	}
}

func TestGetStorageRepositoryByNameLabel(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}
	defaultSr, err := c.GetStorageRepositoryById(accTestPool.DefaultSR)

	if err != nil {
		t.Fatalf("failed to retrieve storage repository by id with error: %v", err)
	}

	srs, err := c.GetStorageRepository(StorageRepository{NameLabel: defaultSr.NameLabel})

	if err != nil {
		t.Fatalf("failed to get storage repository by name label with error: %v", err)
	}

	sr := srs[0]
	if sr.NameLabel != defaultSr.NameLabel {
		t.Errorf("expected storage repository to have name `%s` received `%s` instead.", defaultSr.NameLabel, sr.NameLabel)
	}
}
