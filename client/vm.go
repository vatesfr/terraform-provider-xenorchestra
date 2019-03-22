package client

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
)

type allObjectResponse struct {
	Objects map[string]Vm `json:"-"`
}

type CPUs struct {
	Number int
}

type MemoryObject struct {
	Dynamic []int `json:"dynamic"`
	Static  []int `json:"static"`
	Size    int   `json:"size"`
}

type Vm struct {
	Type               string       `json:"type,omitempty"`
	Id                 string       `json:"id,omitempty"`
	Name               string       `json:"name,omitempty"`
	NameDescription    string       `json:"name_description"`
	NameLabel          string       `json:"name_label"`
	CPUs               CPUs         `json:"CPUs"`
	Memory             MemoryObject `json:"memory"`
	PowerState         string       `json:"power_state"`
	VIFs               []string     `json:"VIFs"`
	VirtualizationMode string       `json:"virtualizationMode"`
	PoolId             string       `json:"$poolId"`
	Template           string       `json:"template"`
	CloudConfig        string       `json:"cloudConfig"`
}

func (c *Client) CreateVm(name_label, name_description, template, cloudConfig string, cpus, memoryMax int, network_ids []string) (*Vm, error) {
	vifs := []map[string]string{}
	for _, network_id := range network_ids {
		vifs = append(vifs, map[string]string{
			"network": network_id,
		})
	}
	params := map[string]interface{}{
		"bootAfterCreate":  true,
		"name_label":       name_label,
		"name_description": name_description,
		"template":         template,
		"cloudConfig":      cloudConfig,
		"coreOs":           false,
		"cpuCap":           nil,
		"cpuWeight":        nil,
		"CPUs":             cpus,
		"memoryMax":        memoryMax,
		"existingDisks": map[string]interface{}{
			"0": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "Created by XO",
				"name_label":       "Ubuntu Bionic Beaver 18.04_imavo",
				"size":             32212254720,
			},
			"1": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "",
				"name_label":       "XO CloudConfigDrive",
				"size":             10485760,
			},
			"2": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "",
				"name_label":       "XO CloudConfigDrive",
				"size":             10485760,
			},
		},
		"VIFs": vifs,
	}
	fmt.Printf("VM params %#v", params)
	var vmId string
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
	err := c.rpc.Call(ctx, "vm.create", params, &vmId)

	if err != nil {
		return nil, err
	}

	err = c.waitForModifyVm(vmId, 5*time.Minute)

	if err != nil {
		return nil, err
	}

	vm := Vm{
		Id: vmId,
	}

	return &vm, nil
}

func (c *Client) DeleteVm(id string) error {
	params := map[string]interface{}{
		"id": id,
	}
	var reply []interface{}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	err := c.rpc.Call(ctx, "vm.delete", params, &reply)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetVm(id string) (*Vm, error) {
	params := map[string]interface{}{
		"type": "VM",
	}
	var objsRes allObjectResponse
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := c.rpc.Call(ctx, "xo.getAllObjects", params, &objsRes.Objects)

	if err != nil {
		return nil, err
	}

	vm, ok := objsRes.Objects[id]
	if !ok {
		return nil, NotFound{
			Id:   id,
			Type: "Vm",
		}
	}
	return &vm, nil
}

func (c *Client) waitForModifyVm(id string, timeout time.Duration) error {
	refreshFn := func() (result interface{}, state string, err error) {
		vm, err := c.GetVm(id)

		if err != nil {
			return vm, "", err
		}

		return vm, vm.PowerState, nil
	}
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Halted", "Stopped"},
		Refresh: refreshFn,
		Target:  []string{"Running"},
		Timeout: timeout,
	}
	_, err := stateConf.WaitForState()
	return err
}
