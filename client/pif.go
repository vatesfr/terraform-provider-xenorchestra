package client

import "errors"

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

func (c *Client) GetPIFByDevice(dev string, vlan int) ([]PIF, error) {
	obj, err := c.FindFromGetAllObjects(PIF{Device: dev, Vlan: vlan})

	if err != nil {
		return []PIF{}, err
	}
	pifs, ok := obj.([]PIF)

	if !ok {
		return pifs, errors.New("failed to coerce response into PIF slice")
	}

	return pifs, nil
}
