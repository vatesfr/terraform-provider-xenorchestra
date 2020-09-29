package client

type PIF struct {
	Device   string
	Host     string
	Network  string
	Id       string
	Uuid     string
	PoolId   string
	Attached bool
	Vlan     int
}

func (p PIF) Compare(obj interface{}) bool {
	otherPif := obj.(PIF)
	if p.Vlan == otherPif.Vlan && p.Device == otherPif.Device {
		return true
	}
	return false
}

func (p PIF) New(obj map[string]interface{}) XoObject {
	id := obj["id"].(string)
	device := obj["device"].(string)
	attached := obj["attached"].(bool)
	network := obj["$network"].(string)
	uuid := obj["uuid"].(string)
	poolId := obj["$poolId"].(string)
	host := obj["$host"].(string)
	vlan := int(obj["vlan"].(float64))
	return PIF{
		Device:   device,
		Host:     host,
		Network:  network,
		Id:       id,
		Uuid:     uuid,
		PoolId:   poolId,
		Attached: attached,
		Vlan:     vlan,
	}
}

func (c *Client) GetPIFByDevice(dev string, vlan int) (PIF, error) {
	obj, err := c.FindFromGetAllObjects(PIF{Device: dev, Vlan: vlan})
	pif := obj.(PIF)

	if err != nil {
		return pif, err
	}

	return pif, nil
}
