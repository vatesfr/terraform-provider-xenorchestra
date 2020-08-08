package client

type StorageRepository struct {
	Id        string
	Uuid      string
	NameLabel string
	PoolId    string
	SRType    string
}

func (s StorageRepository) Compare(obj map[string]interface{}) bool {
	nameLabel := obj["name_label"].(string)
	poolId := obj["$poolId"].(string)
	if s.NameLabel != nameLabel {
		return false
	}

	if s.PoolId == "" {
		return true
	}

	if s.PoolId == poolId {
		return true
	}

	return false
}

func (s StorageRepository) New(obj map[string]interface{}) XoObject {
	id := obj["id"].(string)
	srType := obj["SR_type"].(string)
	poolId := obj["$poolId"].(string)
	nameLabel := obj["name_label"].(string)
	uuid := obj["uuid"].(string)
	return StorageRepository{
		Id:        id,
		NameLabel: nameLabel,
		PoolId:    poolId,
		SRType:    srType,
		Uuid:      uuid,
	}
}

func (c *Client) GetStorageRepository(sr StorageRepository) (StorageRepository, error) {
	obj, err := c.FindFromGetAllObjects(sr)
	sr = obj.(StorageRepository)

	if err != nil {
		return sr, err
	}

	return sr, nil
}
