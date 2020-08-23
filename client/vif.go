package client

type VIF struct {
	Network    string
	Device     string
	MacAddress string
	VmId       string
}

func (v VIF) New(obj map[string]interface{}) XoObject {
	vmId := obj["$VM"].(string)
	device := obj["device"].(string)
	networkId := obj["$network"].(string)
	macAddress := obj["MAC"].(string)
	return VIF{
		Device:     device,
		MacAddress: macAddress,
		Network:    networkId,
		VmId:       vmId,
	}
}

func (v VIF) Compare(obj map[string]interface{}) bool {
	vmId := obj["$VM"].(string)
	if v.VmId != vmId {
		return false
	}
	return true
}

func (c *Client) GetVIFs(vm *Vm) ([]VIF, error) {
	obj, err := c.FindFromGetAllObjects(VIF{VmId: vm.Id})

	if err != nil {
		return nil, err
	}

	if vif, ok := obj.(VIF); ok {
		return []VIF{vif}, nil
	}

	objs := obj.([]interface{})
	vifs := make([]VIF, len(objs))
	for _, vif := range objs {
		vifs = append(vifs, vif.(VIF))
	}
	return vifs, nil
}
