package client

import (
	"errors"
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
	NameDescription    string       `json:"name_description"`
	NameLabel          string       `json:"name_label"`
	CPUs               CPUs         `json:"CPUs"`
	Memory             MemoryObject `json:"memory"`
	PowerState         string       `json:"power_state"`
	VIFs               []string     `json:"VIFs"`
	VBDs               []string     `json:"$VBDs"`
	VirtualizationMode string       `json:"virtualizationMode"`
	PoolId             string       `json:"$poolId"`
	Template           string       `json:"template"`
	AutoPoweron        bool         `json:"auto_poweron"`
	HA                 string       `json:"high_availability"`
	CloudConfig        string       `json:"cloudConfig"`
	ResourceSet        string       `json:"resourceSet,omitempty"`
	Tags               []string     `json:"tags"`
}

func (v Vm) Compare(obj interface{}) bool {
	other := obj.(Vm)
	if v.Id != "" && v.Id == other.Id {
		return true
	}

	if v.NameLabel != "" && v.NameLabel == other.NameLabel {
		return true
	}

	tagCount := len(v.Tags)
	if tagCount > 0 {
		for _, tag := range v.Tags {
			if stringInSlice(tag, other.Tags) {
				tagCount--
			}
		}

		if tagCount == 0 {
			return true
		}
	}

	return false
}

func (c *Client) CreateVm(name_label, name_description, template, cloudConfig, resourceSet string, cpus, memoryMax int, networks []map[string]string, disks []VDI) (*Vm, error) {
	vifs := []map[string]string{}
	for _, network := range networks {
		vifs = append(vifs, map[string]string{
			"network": network["network_id"],
			"mac":     network["mac_address"],
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
		"coreOs":           false,
		"cpuCap":           nil,
		"cpuWeight":        nil,
		"CPUs":             cpus,
		"memoryMax":        memoryMax,
		"existingDisks":    existingDisks,
		"VIFs":             vifs,
	}

	if cloudConfig != "" {
		params["cloudConfig"] = cloudConfig
	}

	if resourceSet != "" {
		params["resourceSet"] = resourceSet
	}
	log.Printf("[DEBUG] VM params for vm.create %#v", params)
	var vmId string
	err := c.Call("vm.create", params, &vmId)

	if err != nil {
		return nil, err
	}

	err = c.waitForModifyVm(vmId, 5*time.Minute)

	if err != nil {
		return nil, err
	}

	return c.GetVm(
		Vm{
			Id: vmId,
		},
	)
}

func (c *Client) UpdateVm(id string, cpus int, nameLabel, nameDescription, ha, rs string, autoPowerOn bool) (*Vm, error) {

	var resourceSet interface{} = rs
	if rs == "" {
		resourceSet = nil
	}
	params := map[string]interface{}{
		"id":                id,
		"name_label":        nameLabel,
		"name_description":  nameDescription,
		"auto_poweron":      autoPowerOn,
		"resourceSet":       resourceSet,
		"high_availability": ha, // valid options are best-effort, restart, ''
		// TODO: VM must be halted in order to change CPUs
		// "CPUs":             cpus,
		// "memoryMax": memoryMax,
		// TODO: These need more investigation before they are implemented
		// pv_args, cpuMask cpuWeight cpuCap affinityHost vga videoram coresPerSocket hasVendorDevice expNestedHvm resourceSet share startDelay nicType hvmBootFirmware virtualizationMode
	}
	log.Printf("[DEBUG] VM params for vm.set: %#v", params)
	var success bool
	err := c.Call("vm.set", params, &success)

	if err != nil {
		return nil, err
	}

	// TODO: This is a poor way to ensure that terraform will see the updated
	// attributes after calling vm.set. Need to investigate a better way to detect this.
	time.Sleep(25 * time.Second)

	return c.GetVm(Vm{Id: id})
}

func (c *Client) DeleteVm(id string) error {
	params := map[string]interface{}{
		"id": id,
	}
	var reply []interface{}
	err := c.Call("vm.delete", params, &reply)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetVm(vmReq Vm) (*Vm, error) {
	obj, err := c.FindFromGetAllObjects(vmReq)

	if err != nil {
		return nil, err
	}
	vms := obj.([]Vm)

	if len(vms) != 1 {
		return nil, errors.New(fmt.Sprintf("expected to find a single VM from request %+v, instead found %d", vmReq, len(vms)))
	}

	log.Printf("[DEBUG] Found vm: %+v", vms[0])
	return &vms[0], nil
}

func (c *Client) GetVms() ([]Vm, error) {
	var response map[string]Vm
	err := c.GetAllObjectsOfType(Vm{PowerState: "Running"}, &response)

	if err != nil {
		return []Vm{}, err
	}

	vms := make([]Vm, 0, len(response))
	for _, vm := range response {
		vms = append(vms, vm)
	}

	log.Printf("[DEBUG] Found vms: %+v", vms)
	return vms, nil
}

func (c *Client) waitForModifyVm(id string, timeout time.Duration) error {
	refreshFn := func() (result interface{}, state string, err error) {
		vm, err := c.GetVm(Vm{Id: id})

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

func FindVmForTests(vm *Vm, tag string) error {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	vmRes, err := c.GetVm(Vm{
		Tags: []string{tag},
	})

	if err != nil {
		return err
	}

	*vm = *vmRes

	return nil
}
