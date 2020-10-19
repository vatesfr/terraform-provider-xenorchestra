package client

import (
	"errors"
	"fmt"
	"os"
)

type StorageRepository struct {
	Id        string   `json:"id"`
	Uuid      string   `json:"uuid"`
	NameLabel string   `json:"name_label"`
	PoolId    string   `json:"$poolId"`
	SRType    string   `json:"SR_type"`
	Tags      []string `json:"tags,omitempty"`
}

func (s StorageRepository) Compare(obj interface{}) bool {
	otherSr := obj.(StorageRepository)

	if s.Id != "" && s.Id == otherSr.Id {
		return true
	}

	if len(s.Tags) > 0 {
		for _, tag := range s.Tags {
			if !stringInSlice(tag, otherSr.Tags) {
				return false
			}
		}
	}

	labelsMatch := false
	if s.NameLabel == otherSr.NameLabel {
		labelsMatch = true
	}

	if s.PoolId == "" && labelsMatch {
		return true
	} else if s.PoolId == otherSr.PoolId && labelsMatch {
		return true
	}

	return false
}

func stringInSlice(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func (c *Client) GetStorageRepositoryById(id string) (StorageRepository, error) {
	obj, err := c.FindFromGetAllObjects(StorageRepository{Id: id})
	var sr StorageRepository

	if err != nil {
		return sr, err
	}
	srs, ok := obj.([]StorageRepository)

	if !ok {
		return sr, errors.New("failed to coerce response into StorageRepository slice")
	}

	if len(srs) != 1 {
		return sr, errors.New(fmt.Sprintf("expected a single storage respository to be returned, instead received: %d in the response: %v", len(srs), obj))
	}

	return srs[0], nil
}

func (c *Client) GetStorageRepository(sr StorageRepository) ([]StorageRepository, error) {
	obj, err := c.FindFromGetAllObjects(sr)

	if err != nil {
		return nil, err
	}
	srs, ok := obj.([]StorageRepository)

	if !ok {
		return nil, errors.New("failed to coerce response into StorageRepository slice")
	}

	return srs, nil
}

func FindStorageRepositoryForTests(pool Pool, sr *StorageRepository, tag string) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	defaultSr, err := c.GetStorageRepositoryById(pool.DefaultSR)

	if err != nil {
		fmt.Printf("failed to find the default storage repository with id: %s with error: %v\n", pool.DefaultSR, err)
		os.Exit(-1)
	}

	*sr = defaultSr

	err = c.AddTag(defaultSr.Id, tag)

	if err != nil {
		fmt.Printf("failed to set tag on default storage repository with id: %s with error: %v\n", pool.DefaultSR, err)
		os.Exit(-1)
	}
}
