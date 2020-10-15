package client

import (
	"errors"
)

type PIF struct {
	Device   string `json:"device"`
	Host     string `json:"$host"`
	Network  string `json:"$network"`
	Id       string `json:"id"`
	Uuid     string `json:"uuid"`
	PoolId   string `json:"$poolId"`
	Attached bool   `json:"attached"`
	Vlan     int    `json:"vlan"`
}

func (p PIF) Compare(obj interface{}) bool {
	otherPif := obj.(PIF)

	hostIdExists := p.Host != ""
	if hostIdExists && p.Host != otherPif.Host {
		return false
	}

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

func (c *Client) GetPIF(pifReq PIF) (pifs []PIF, err error) {
	obj, err := c.FindFromGetAllObjects(pifReq)

	if err != nil {
		return
	}
	pifs, ok := obj.([]PIF)

	if !ok {
		return pifs, errors.New("failed to coerce response into PIF slice")
	}

	return pifs, nil
}
