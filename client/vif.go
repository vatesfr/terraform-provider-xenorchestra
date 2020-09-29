package client

import (
	"context"
	"log"
	"time"
)

type VIF struct {
	Id         string
	Attached   bool
	Network    string
	Device     string
	MacAddress string
	VmId       string
}

func (v VIF) New(obj map[string]interface{}) XoObject {
	id := obj["id"].(string)
	attached := obj["attached"].(bool)
	vmId := obj["$VM"].(string)
	device := obj["device"].(string)
	networkId := obj["$network"].(string)
	macAddress := obj["MAC"].(string)
	return VIF{
		Id:         id,
		Attached:   attached,
		Device:     device,
		MacAddress: macAddress,
		Network:    networkId,
		VmId:       vmId,
	}
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

	obj, err := c.FindFromGetAllObjects(VIF{
		Id:         vifReq.Id,
		MacAddress: vifReq.MacAddress,
	})

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

	params := map[string]interface{}{
		"id": vif.Id,
	}
	var result bool
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = c.Call(ctx, "vif.disconnect", params, &result)
	log.Printf("[DEBUG] Calling vif.disconnect received err: %v", err)

	if err != nil {
		return err
	}

	err = c.Call(ctx, "vif.delete", params, &result)
	log.Printf("[DEBUG] Calling vif.delete received err: %v", err)

	if err != nil {
		return err
	}

	return nil
}
