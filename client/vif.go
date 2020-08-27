package client

import (
	"context"
	"time"
)

type VIF struct {
	Id         string
	Network    string
	Device     string
	MacAddress string
	VmId       string
}

func (v VIF) New(obj map[string]interface{}) XoObject {
	id := obj["id"].(string)
	vmId := obj["$VM"].(string)
	device := obj["device"].(string)
	networkId := obj["$network"].(string)
	macAddress := obj["MAC"].(string)
	return VIF{
		Id:         id,
		Device:     device,
		MacAddress: macAddress,
		Network:    networkId,
		VmId:       vmId,
	}
}

func (v VIF) Compare(obj map[string]interface{}) bool {
	id := obj["id"].(string)
	if v.Id == id {
		return true
	}

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
	var vifs []VIF
	for _, vif := range objs {
		vifs = append(vifs, vif.(VIF))
	}
	return vifs, nil
}

func (c *Client) GetVIF(vifReq *VIF) (*VIF, error) {

	obj, err := c.FindFromGetAllObjects(VIF{Id: vifReq.Id})

	if err != nil {
		return nil, err
	}

	vif := obj.(VIF)
	return &vif, nil
}

func (c *Client) CreateVIF(vm *Vm, vif *VIF) (*VIF, error) {

	var id string
	params := map[string]interface{}{
		"network": vif.Network,
		"vm":      vm.Id,
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := c.Call(ctx, "vm.createInterface", params, &id)

	if err != nil {
		return nil, err
	}

	return c.GetVIF(&VIF{Id: id})
}

func (c *Client) DeleteVIF(vif *VIF) error {
	var result bool
	params := map[string]interface{}{
		"id": vif.Id,
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := c.Call(ctx, "vif.disconnect", params, &result)

	if err != nil {
		return err
	}

	err = c.Call(ctx, "vif.delete", params, &result)

	if err != nil {
		return err
	}

	return nil
}
