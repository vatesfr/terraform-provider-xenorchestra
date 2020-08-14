package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
	AutoPoweron        bool         `json:"auto_poweron"`
	HA                 string       `json:"high_availability"`
	CloudConfig        string       `json:"cloudConfig"`
}

type VDI struct {
	SrId      string
	NameLabel string
	Size      int
}

func (c *Client) CreateVm(name_label, name_description, template, cloudConfig string, cpus, memoryMax int, network_ids []string, disks []VDI) (*Vm, error) {
	vifs := []map[string]string{}
	for _, network_id := range network_ids {
		vifs = append(vifs, map[string]string{
			"network": network_id,
		})
	}
	existingDisks := map[string]interface{}{}

	for idx, disk := range disks {
		existingDisks[fmt.Sprintf("%d", idx)] = map[string]interface{}{
			"$SR":        disk.SrId,
			"name_label": disk.NameLabel,
			"size":       disk.Size,
		}
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
		"existingDisks":    existingDisks,
		"VIFs":             vifs,
	}
	log.Printf("[DEBUG] VM params for vm.create %#v", params)
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

func (c *Client) UpdateVm(id string, cpus int, nameLabel, nameDescription, ha string, autoPowerOn bool) (*Vm, error) {
	params := map[string]interface{}{
		"id":                id,
		"name_label":        nameLabel,
		"name_description":  nameDescription,
		"auto_poweron":      autoPowerOn,
		"high_availability": ha, // valid options are best-effort, restart, ''
		// TODO: VM must be halted in order to change CPUs
		// "CPUs":             cpus,
		// "memoryMax": memoryMax,
		// TODO: These need more investigation before they are implemented
		// pv_args, cpuMask cpuWeight cpuCap affinityHost vga videoram coresPerSocket hasVendorDevice expNestedHvm resourceSet share startDelay nicType hvmBootFirmware virtualizationMode
	}
	log.Printf("[DEBUG] VM params for vm.set: %#v", params)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
	var success bool
	err := c.rpc.Call(ctx, "vm.set", params, &success)

	if err != nil {
		return nil, err
	}

	// TODO: This is a poor way to ensure that terraform will see the updated
	// attributes after calling vm.set. Need to investigate a better way to detect this.
	time.Sleep(time.Minute)

	return c.GetVm(id)
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
		"filter": map[string]string{
			"type": "VM",
		},
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
	log.Printf("[DEBUG] Found vm: %+v", vm)
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
