package client

import (
	"testing"
)

func TestGetStorageRepositoryByType(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client with error: %v", err)
	}
	expectedNameLabel := "XenServer Tools"
	sr, err := c.GetStorageRepositoryByName(expectedNameLabel)

	if err != nil {
		t.Errorf("failed to get storage repository with error: %v", err)
	}

	if sr.NameLabel != expectedNameLabel {
		t.Errorf("expected storage repository to have name `%s` received `%s` instead.", expectedNameLabel, sr.NameLabel)
	}
}
