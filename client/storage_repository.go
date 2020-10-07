package client

import (
	"errors"
	"fmt"
)

type StorageRepository struct {
	Id        string `json:"id"`
	Uuid      string `json:"uuid"`
	NameLabel string `json:"name_label"`
	PoolId    string `json:"$poolId"`
	SRType    string `json:"SR_type"`
}

func (s StorageRepository) Compare(obj interface{}) bool {
	otherSr := obj.(StorageRepository)

	if s.Id == otherSr.Id {
		return true
	}

	if s.NameLabel == otherSr.NameLabel {
		return true
	}

	if s.PoolId == otherSr.PoolId {
		return true
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
