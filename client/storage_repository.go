package client

type StorageRepository struct {
	Id       string
	Uuid     string
	PoolId   string
	Attached bool
	Vlan     int
}

func (s StorageRepository) Compare(obj map[string]interface{}) bool {
	// device := obj["device"].(string)
	// vlan := int(obj["vlan"].(float64))
	// if p.Vlan == vlan && p.Device == device {
	// 	return true
	// }
	return false
}

func (s StorageRepository) New(obj map[string]interface{}) XoObject {
	id := obj["id"].(string)
	return StorageRepository{
		Id: id,
	}
}

func (c *Client) GetStorageRepositoryByType(srType string, pool string) (PIF, error) {
	obj, err := c.FindFromGetAllObjects(PIF{Device: dev, Vlan: vlan})
	sr := obj.(StorageRepository)

	if err != nil {
		return sr, err
	}

	return sr, nil
}
