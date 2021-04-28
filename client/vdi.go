package client

import (
	"errors"
	"fmt"
)

type Disk struct {
	VBD
	VDI
}

type VDI struct {
	VDIId           string   `json:"id"`
	SrId            string   `json:"$SR"`
	NameLabel       string   `json:"name_label"`
	NameDescription string   `json:"name_description"`
	Size            int      `json:"size"`
	VBDs            []string `json:"$VBDs"`
	PoolId          string   `json:"$poolId"`
	Tags            []string `json:"tags,omitempty"`
}

func (v VDI) Compare(obj interface{}) bool {
	other := obj.(VDI)

	if v.VDIId != "" && other.VDIId == v.VDIId {
		return true
	}

	labelsMatch := false
	if v.NameLabel == other.NameLabel {
		labelsMatch = true
	}

	if v.PoolId == other.PoolId && labelsMatch {
		return true
	}

	if len(v.Tags) > 0 {
		for _, tag := range v.Tags {
			if !stringInSlice(tag, other.Tags) {
				return false
			}
		}
	}

	return false
}

// TODO: Change this file to storage or disks?
type VBD struct {
	Id        string `json:"id"`
	Attached  bool
	Device    string
	ReadOnly  bool   `json:"read_only"`
	VmId      string `json:"VM"`
	VDI       string `json:"VDI"`
	IsCdDrive bool   `json:"is_cd_drive"`
	Position  string
	Bootable  bool
	PoolId    string `json:"$poolId"`
}

func (v VBD) Compare(obj interface{}) bool {
	other := obj.(VBD)
	if v.IsCdDrive != other.IsCdDrive {
		return false
	}

	if other.VmId != "" && v.VmId == other.VmId {
		return true
	}

	return false
}

func (c *Client) getDisksFromVBDs(vbd VBD) ([]Disk, error) {
	obj, err := c.FindFromGetAllObjects(vbd)

	if _, ok := err.(NotFound); ok {
		return []Disk{}, nil
	}

	if err != nil {
		return nil, err
	}
	disks, ok := obj.([]VBD)

	if !ok {
		return []Disk{}, errors.New(fmt.Sprintf("failed to coerce %v into VBD", obj))
	}

	vdis := []Disk{}
	for _, disk := range disks {
		vdi, err := c.GetParentVDI(disk)

		if err != nil {
			return []Disk{}, err
		}

		vdis = append(vdis, Disk{disk, vdi})
	}
	return vdis, nil
}

func (c *Client) GetDisks(vm *Vm) ([]Disk, error) {
	return c.getDisksFromVBDs(VBD{
		VmId:      vm.Id,
		IsCdDrive: false,
	})
}

func (c *Client) GetCdroms(vm *Vm) ([]Disk, error) {
	cds, err := c.getDisksFromVBDs(VBD{
		VmId:      vm.Id,
		IsCdDrive: true,
	})

	// Not every Vm will have CDs. Rather than pass
	// this to the caller, catch it and return empty
	// CDs.
	if _, ok := err.(NotFound); ok {
		return []Disk{}, nil
	}

	return cds, err
}

func (c *Client) GetVDIs(vdiReq VDI) ([]VDI, error) {
	obj, err := c.FindFromGetAllObjects(vdiReq)

	if err != nil {
		return nil, err
	}

	vdis, ok := obj.([]VDI)

	if !ok {
		return nil, errors.New(fmt.Sprintf("failed to coerce %+v into VDI", obj))
	}

	return vdis, nil
}

func (c *Client) GetParentVDI(vbd VBD) (VDI, error) {
	obj, err := c.FindFromGetAllObjects(VDI{
		VDIId: vbd.VDI,
	})

	// Rather than detect not found errors for finding the
	// parent VDI this is considered an error so we return
	// it to the caller.
	if err != nil {
		return VDI{}, err
	}
	disks, ok := obj.([]VDI)

	if !ok {
		return VDI{}, errors.New(fmt.Sprintf("failed to coerce %+v into VDI", obj))
	}

	if len(disks) != 1 {
		return VDI{}, errors.New(fmt.Sprintf("expected Vm VDI '%s' to only contain a single VBD, instead found %d: %+v", vbd.VDI, len(disks), disks))
	}
	return disks[0], nil
}

func (c *Client) CreateDisk(vm Vm, d Disk) (string, error) {
	var id string
	params := map[string]interface{}{
		"name": d.NameLabel,
		"size": d.Size,
		"sr":   d.SrId,
		"vm":   vm.Id,
	}
	err := c.Call("disk.create", params, &id)

	return id, err
}

func (c *Client) DeleteDisk(vm Vm, d Disk) error {
	var success bool
	disconnectParams := map[string]interface{}{
		"id": d.Id,
	}
	err := c.Call("vbd.disconnect", disconnectParams, &success)

	if err != nil {
		return err
	}

	vdiDeleteParams := map[string]interface{}{
		"id": d.VDIId,
	}
	return c.Call("vdi.delete", vdiDeleteParams, &success)
}

func (c *Client) ConnectDisk(d Disk) error {
	var success bool
	params := map[string]interface{}{
		"id": d.Id,
	}
	return c.Call("vbd.connect", params, &success)
}

func (c *Client) DisconnectDisk(d Disk) error {
	var success bool
	params := map[string]interface{}{
		"id": d.Id,
	}
	return c.Call("vbd.disconnect", params, &success)
}

func (c *Client) UpdateVDI(d Disk) error {
	var success bool
	params := map[string]interface{}{
		"id":               d.VDIId,
		"name_description": d.NameDescription,
		"name_label":       d.NameLabel,
	}
	return c.Call("vdi.set", params, &success)
}

func (c *Client) EjectCd(id string) error {
	var success bool
	params := map[string]interface{}{
		"id": id,
	}
	return c.Call("vm.ejectCd", params, &success)
}

func (c *Client) InsertCd(vmId, cdId string) error {
	var success bool
	params := map[string]interface{}{
		"id":    vmId,
		"cd_id": cdId,
	}
	return c.Call("vm.insertCd", params, &success)
}
