package client

import (
	"errors"
	"fmt"
	"log"
)

type VIF struct {
	Id         string `json:"id"`
	Attached   bool   `json:"attached"`
	Network    string `json:"$network"`
	Device     string `json:"device"`
	MacAddress string `json:"MAC"`
	VmId       string `json:"$VM"`
}

func (v VIF) Compare(obj interface{}) bool {
	other := obj.(VIF)
	if v.Id == other.Id {
		return true
	}

	if v.MacAddress == other.MacAddress {
		return true
	}

	if v.VmId == other.VmId {
		return true
	}
	return false
}

func (c *Client) GetVIFs(vm *Vm) ([]VIF, error) {
	obj, err := c.FindFromGetAllObjects(VIF{VmId: vm.Id})

	if _, ok := err.(NotFound); ok {
		return []VIF{}, nil
	}

	if err != nil {
		return nil, err
	}
	vifs, ok := obj.([]VIF)
	if !ok {
		return []VIF{}, errors.New("failed to coerce response into VIF slice")
	}

	return vifs, nil
}

func (c *Client) GetVIF(vifReq *VIF) (*VIF, error) {

	obj, err := c.FindFromGetAllObjects(VIF{
		Id:         vifReq.Id,
		MacAddress: vifReq.MacAddress,
	})

	if err != nil {
		return nil, err
	}

	vifs := obj.([]VIF)

	if len(vifs) > 1 {
		return nil, errors.New(fmt.Sprintf("recieved %d VIFs but was expecting a single VIF to be returned", len(vifs)))
	}
	return &vifs[0], nil
}

func (c *Client) CreateVIF(vm *Vm, vif *VIF) (*VIF, error) {

	var id string
	params := map[string]interface{}{
		"network": vif.Network,
		"vm":      vm.Id,
	}
	if vif.MacAddress != "" {
		params["mac"] = vif.MacAddress
	}
	err := c.Call("vm.createInterface", params, &id)

	if err != nil {
		return nil, err
	}

	return c.GetVIF(&VIF{Id: id})
}

func (c *Client) ConnectVIF(vifReq *VIF) (err error) {
	vif, err := c.GetVIF(vifReq)

	if err != nil {
		return
	}
	var success bool
	err = c.Call("vif.connect", map[string]interface{}{
		"id": vif.Id,
	}, &success)
	return
}

func (c *Client) DisconnectVIF(vifReq *VIF) (err error) {
	vif, err := c.GetVIF(vifReq)

	if err != nil {
		return
	}

	var success bool
	err = c.Call("vif.disconnect", map[string]interface{}{
		"id": vif.Id,
	}, &success)
	return
}

func (c *Client) DeleteVIF(vifReq *VIF) (err error) {
	var vif *VIF

	// This is a request that is looking the VIF
	// up by macaddress and needs to lookup the ID first.
	if vifReq.Id == "" {
		vif, err = c.GetVIF(vifReq)

		if err != nil {
			return err
		}
	} else {
		vif = vifReq
	}

	err = c.DisconnectVIF(vif)

	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"id": vif.Id,
	}
	var result bool
	err = c.Call("vif.delete", params, &result)
	log.Printf("[DEBUG] Calling vif.delete received err: %v", err)

	if err != nil {
		return err
	}

	return nil
}
