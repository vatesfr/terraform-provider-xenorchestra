package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type allObjectResponse struct {
	Objects map[string]Vm `json:"-"`
}

const (
	HaltedPowerState    string = "Halted"
	PausedPowerState    string = "Paused"
	RunningPowerState   string = "Running"
	SuspendedPowerState string = "Suspended"
)

const (
	CloneTypeFastClone string = "fast"
	CloneTypeFullClone string = "full"
)

type CPUs struct {
	Number int `json:"number"`
	Max    int `json:"max"`
}

type MemoryObject struct {
	Dynamic []int `json:"dynamic"`
	Static  []int `json:"static"`
	Size    int   `json:"size"`
}

type Boot struct {
	Firmware string `json:"firmware,omitempty"`
}

// The XO api sometimes returns the videoram field as an int
// and sometimes as a string. This overrides the default json
// unmarshalling so that we can handle both of these cases
type Videoram struct {
	Value int `json:"-"`
}

func (v *Videoram) UnmarshalJSON(data []byte) (err error) {
	s := string(data)
	l := len(s)
	if s[0] == '"' && s[l-1] == '"' {
		num := 0
		if l > 2 {
			num, err = strconv.Atoi(s[1 : l-1])

			if err != nil {
				return err
			}

		}
		v.Value = num
		return nil
	}

	return json.Unmarshal(data, &v.Value)
}

// This resource set type is used to allow differentiating between when a
// user wants to remove a resource set from a VM (during an update) and when
// a resource set parameter should be omitted from a vm.set RPC call.
type FlatResourceSet struct {
	Id string
}

// This ensures when a FlatResourceSet is printed in debug logs
// that the string value of the Id is used rather than the pointer
// value. Since the purpose of this struct is to flatten resource
// sets to a string, it makes the logs properly reflect what is
// being logged.
func (rs *FlatResourceSet) String() string {
	return rs.Id
}

func (rs *FlatResourceSet) UnmarshalJSON(data []byte) (err error) {
	return json.Unmarshal(data, &rs.Id)
}

func (rs *FlatResourceSet) MarshalJSON() ([]byte, error) {
	if len(rs.Id) == 0 {
		var buf bytes.Buffer
		buf.WriteString(`null`)
		return buf.Bytes(), nil
	} else {
		return json.Marshal(rs.Id)
	}
}

type Vm struct {
	Addresses          map[string]string `json:"addresses,omitempty"`
	BlockedOperations  map[string]string `json:"blockedOperations,omitempty"`
	Boot               Boot              `json:"boot,omitempty"`
	Type               string            `json:"type,omitempty"`
	Id                 string            `json:"id,omitempty"`
	AffinityHost       *string           `json:"affinityHost,omitempty"`
	NameDescription    string            `json:"name_description"`
	NameLabel          string            `json:"name_label"`
	CPUs               CPUs              `json:"CPUs"`
	ExpNestedHvm       bool              `json:"expNestedHvm,omitempty"`
	Memory             MemoryObject      `json:"memory"`
	PowerState         string            `json:"power_state"`
	VIFs               []string          `json:"VIFs"`
	VBDs               []string          `json:"$VBDs"`
	VirtualizationMode string            `json:"virtualizationMode"`
	PoolId             string            `json:"$poolId"`
	Template           string            `json:"template"`
	AutoPoweron        bool              `json:"auto_poweron"`
	HA                 string            `json:"high_availability"`
	CloudConfig        string            `json:"cloudConfig"`
	ResourceSet        *FlatResourceSet  `json:"resourceSet"`
	// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
	// SecureBoot         bool              `json:"secureBoot,omitempty"`
	Tags         []string               `json:"tags"`
	Videoram     Videoram               `json:"videoram,omitempty"`
	Vga          string                 `json:"vga,omitempty"`
	StartDelay   int                    `json:startDelay,omitempty"`
	Host         string                 `json:"$container"`
	XenstoreData map[string]interface{} `json:"xenStoreData,omitempty"`

	// These fields are used for passing in disk inputs when
	// creating Vms, however, this is not a real field as far
	// as the XO api or XAPI is concerned
	Disks              []Disk              `json:"-"`
	CloudNetworkConfig string              `json:"-"`
	VIFsMap            []map[string]string `json:"-"`
	// Map where the key is the network interface index and the
	// value is a cidr range parsable by net.ParseCIDR
	WaitForIps                     map[string]string `json:"-"`
	Installation                   Installation      `json:"-"`
	ManagementAgentDetected        bool              `json:"managementAgentDetected"`
	PVDriversDetected              bool              `json:"pvDriversDetected"`
	DestroyCloudConfigVdiAfterBoot bool              `json:"-"`
	CloneType                      string            `json:"-"`
}

type Installation struct {
	Method     string `json:"-"`
	Repository string `json:"-"`
}

func (v Vm) Compare(obj interface{}) bool {
	other := obj.(Vm)
	if v.Id != "" && v.Id == other.Id {
		return true
	}

	if v.NameLabel != "" && v.NameLabel == other.NameLabel {
		return true
	}
	if v.PowerState != "" && v.Host != "" {
		if (v.PowerState == other.PowerState) && (v.Host == other.Host) {
			return true
		}
		return false
	} else if v.PowerState != "" && v.PowerState == other.PowerState {
		return true
	} else if v.Host != "" && v.Host == other.Host {
		return true
	}
	if v.PoolId != "" && v.PoolId == other.PoolId {
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

func (c *Client) SuspendVm(id string) error {
	return c.changeVmState(id, "suspend", []string{SuspendedPowerState}, []string{RunningPowerState}, 2*time.Minute)
}

func (c *Client) changeVmState(id, action string, target, pending []string, timeout time.Duration) error {
	// PV drivers are necessary for the XO api to issue a graceful shutdown.
	// See https://github.com/terra-farm/terraform-provider-xenorchestra/issues/220
	// for more details.
	if err := c.waitForPVDriversDetected(id); err != nil {
		return errors.New(
			fmt.Sprintf("failed to gracefully %s vm (%s) since PV drivers were never detected", action, id))
	}

	params := map[string]interface{}{
		"id": id,
	}
	var success bool
	err := c.Call(fmt.Sprintf("vm.%s", action), params, &success)

	if err != nil {
		return err
	}
	return c.waitForVmState(
		id,
		StateChangeConf{
			Pending: pending,
			Target:  target,
			Timeout: timeout,
		},
	)
}

func (c *Client) PauseVm(id string) error {
	return c.changeVmState(id, "pause", []string{PausedPowerState}, []string{RunningPowerState}, 2*time.Minute)
}

func (c *Client) CreateVm(vmReq Vm, createTime time.Duration) (*Vm, error) {
	tmpl, err := c.GetTemplate(Template{
		Id: vmReq.Template,
	})

	if err != nil {
		return nil, err
	}

	if len(tmpl) != 1 {
		return nil, errors.New(fmt.Sprintf("cannot create VM when multiple templates are returned: %v", tmpl))
	}

	useExistingDisks := tmpl[0].isDiskTemplate()
	installation := vmReq.Installation
	if !useExistingDisks && installation.Method != "cdrom" && installation.Method != "network" {
		return nil, errors.New("cannot create a VM from a diskless template without an ISO")
	}

	existingDisks := map[string]interface{}{}
	vdis := []interface{}{}
	disks := vmReq.Disks

	firstDisk := createVdiMap(disks[0])
	// Treat the first disk differently. This covers the
	// case where we are using a template with an already
	// installed OS or a diskless template.
	if useExistingDisks {
		existingDisks["0"] = firstDisk
	} else {
		vdis = append(vdis, firstDisk)
	}

	for i := 1; i < len(disks); i++ {
		vdis = append(vdis, createVdiMap(disks[i]))
	}

	params := map[string]interface{}{
		"bootAfterCreate":  false,
		"clone":            useExistingDisks && vmReq.CloneType == CloneTypeFastClone,
		"name_label":       vmReq.NameLabel,
		"name_description": vmReq.NameDescription,
		"template":         vmReq.Template,
		"coreOs":           false,
		"cpuCap":           nil,
		"cpuWeight":        nil,
		"CPUs":             vmReq.CPUs.Number,
		"memoryMax":        vmReq.Memory.Static[1],
		"existingDisks":    existingDisks,
		// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
		// "secureBoot":       vmReq.SecureBoot,
		"expNestedHvm":      vmReq.ExpNestedHvm,
		"VDIs":              vdis,
		"VIFs":              vmReq.VIFsMap,
		"tags":              vmReq.Tags,
		"auto_poweron":      vmReq.AutoPoweron,
		"high_availability": vmReq.HA,
	}

	if !params["clone"].(bool) && vmReq.CloneType == CloneTypeFastClone {
		fmt.Printf("[WARN] A fast clone was requested but falling back to full due to lack of disk template support\n")
	}

	destroyCloudConfigVdiAfterBoot := vmReq.DestroyCloudConfigVdiAfterBoot
	if destroyCloudConfigVdiAfterBoot {
		params["destroyCloudConfigVdiAfterBoot"] = destroyCloudConfigVdiAfterBoot
		params["bootAfterCreate"] = true
	}

	videoram := vmReq.Videoram.Value
	if videoram != 0 {
		params["videoram"] = videoram
	}

	firmware := vmReq.Boot.Firmware
	if firmware != "" {
		params["hvmBootFirmware"] = firmware
	}

	affinityHost := vmReq.AffinityHost
	if affinityHost != nil {
		params["affinityHost"] = affinityHost
	}

	vga := vmReq.Vga
	if vga != "" {
		params["vga"] = vga
	}

	startDelay := vmReq.StartDelay
	if startDelay != 0 {
		params["startDelay"] = startDelay
	}

	if len(vmReq.BlockedOperations) > 0 {
		blockedOperations := map[string]string{}
		for _, v := range vmReq.BlockedOperations {
			blockedOperations[v] = "true"
		}
		params["blockedOperations"] = blockedOperations
	}

	if installation.Method != "" {
		params["installation"] = map[string]string{
			"method":     installation.Method,
			"repository": installation.Repository,
		}
	}

	cloudConfig := vmReq.CloudConfig
	if cloudConfig != "" {
		warnOnInvalidCloudConfig(cloudConfig)

		params["cloudConfig"] = cloudConfig
	}

	resourceSet := vmReq.ResourceSet
	if resourceSet != nil {
		params["resourceSet"] = resourceSet
	}

	cloudNetworkConfig := vmReq.CloudNetworkConfig
	if cloudNetworkConfig != "" {
		params["networkConfig"] = cloudNetworkConfig
	}
	log.Printf("[DEBUG] VM params for vm.create %#v", params)
	var vmId string
	err = c.Call("vm.create", params, &vmId)

	if err != nil {
		return nil, err
	}

	xsParams := map[string]interface{}{
		"id":           vmId,
		"xenStoreData": vmReq.XenstoreData,
	}
	var success bool
	err = c.Call("vm.set", xsParams, &success)

	if err != nil {
		return nil, err
	}

	bootAfterCreate := params["bootAfterCreate"].(bool)
	if !bootAfterCreate && vmReq.PowerState == RunningPowerState {
		err = c.StartVm(vmId)
		if err != nil {
			return nil, err
		}
	}

	err = c.waitForModifyVm(vmId, vmReq.PowerState, vmReq.WaitForIps, createTime)

	if err != nil {
		return nil, err
	}

	return c.GetVm(
		Vm{
			Id: vmId,
		},
	)
}

func createVdiMap(disk Disk) map[string]interface{} {
	return map[string]interface{}{
		"$SR":              disk.SrId,
		"SR":               disk.SrId,
		"name_label":       disk.NameLabel,
		"name_description": disk.NameDescription,
		"size":             disk.Size,
		"type":             "user",
	}
}

func (c *Client) UpdateVm(vmReq Vm) (*Vm, error) {
	params := map[string]interface{}{
		"id":                vmReq.Id,
		"name_label":        vmReq.NameLabel,
		"name_description":  vmReq.NameDescription,
		"auto_poweron":      vmReq.AutoPoweron,
		"high_availability": vmReq.HA, // valid options are best-effort, restart, ''
		"CPUs":              vmReq.CPUs.Number,
		"memoryMax":         vmReq.Memory.Static[1],
		"expNestedHvm":      vmReq.ExpNestedHvm,
		"startDelay":        vmReq.StartDelay,
		// TODO: These need more investigation before they are implemented
		// pv_args

		// virtualizationMode hvm or pv, cannot be set after vm is created (requires conversion)

		// hasVendorDevice must be applied when the vm is halted and only applies to windows machines - https://github.com/xapi-project/xen-api/blob/889b83c47d46c4df65fe58b01caed284dab8dc93/ocaml/idl/datamodel_vm.ml#L1168

		// share relates to resource sets. This can be accomplished with the resource set resource so supporting it isn't necessary

		// cpusMask, cpuWeight and cpuCap can be changed at runtime to an integer value or null
		// coresPerSocket is null or a number of cores per socket. Putting an invalid value doesn't seem to cause an error :(
	}

	affinityHost := vmReq.AffinityHost
	if affinityHost != nil {
		if *affinityHost == "" {
			params["affinityHost"] = nil
		} else {
			params["affinityHost"] = *affinityHost
		}
	}

	videoram := vmReq.Videoram.Value
	if videoram != 0 {
		params["videoram"] = videoram
	}

	if vmReq.ResourceSet != nil {
		params["resourceSet"] = vmReq.ResourceSet
	}

	if len(vmReq.XenstoreData) > 0 {
		params["xenStoreData"] = vmReq.XenstoreData
	}

	vga := vmReq.Vga
	if vga != "" {
		params["vga"] = vga
	}

	// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
	// secureBoot := vmReq.SecureBoot
	// if secureBoot {
	// 	params["secureBoot"] = true
	// }

	firmware := vmReq.Boot.Firmware
	if firmware != "" {
		params["hvmBootFirmware"] = firmware
	}

	blockedOperations := map[string]interface{}{}
	for k, v := range vmReq.BlockedOperations {
		if v == "false" {
			blockedOperations[k] = nil

		} else {
			blockedOperations[k] = v
		}
	}
	params["blockedOperations"] = blockedOperations

	log.Printf("[DEBUG] VM params for vm.set: %#v", params)

	var success bool
	err := c.Call("vm.set", params, &success)

	if err != nil {
		return nil, err
	}

	// TODO: This is a poor way to ensure that terraform will see the updated
	// attributes after calling vm.set. Need to investigate a better way to detect this.
	time.Sleep(25 * time.Second)

	return c.GetVm(vmReq)
}

func (c *Client) StartVm(id string) error {
	params := map[string]interface{}{
		"id": id,
	}
	var success bool
	// TODO: This can block indefinitely before we get to the waitForVmHalt
	err := c.Call("vm.start", params, &success)

	if err != nil {
		return err
	}
	return c.waitForVmState(
		id,
		StateChangeConf{
			Pending: []string{HaltedPowerState},
			Target:  []string{RunningPowerState},
			Timeout: 2 * time.Minute,
		},
	)
}

func (c *Client) HaltVm(id string) error {
	return c.changeVmState(id, "stop", []string{HaltedPowerState}, []string{RunningPowerState}, 2*time.Minute)
}

func (c *Client) DeleteVm(id string) error {
	params := map[string]interface{}{
		"id": id,
	}
	// Xen Orchestra versions >= 5.69.0 changed this return value to a bool
	// when older versions returned an object. This needs to be an interface
	// type in order to be backwards compatible while fixing this bug. See
	// GitHub issue 196 for more details.
	var reply interface{}
	return c.Call("vm.delete", params, &reply)
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

func (c *Client) GetVms(vm Vm) ([]Vm, error) {
	obj, err := c.FindFromGetAllObjects(vm)
	if err != nil {
		return []Vm{}, err
	}
	vms := obj.([]Vm)
	log.Printf("[DEBUG] Found vms: %+v", vms)
	return vms, nil
}

func (c *Client) EjectVmCd(vm *Vm) error {
	params := map[string]interface{}{
		"id": vm.Id,
	}
	var result bool
	err := c.Call("vm.ejectCd", params, &result)
	if err != nil || !result {
		return err
	}
	return nil
}

func GetVmPowerState(c *Client, id string) func() (result interface{}, state string, err error) {
	return func() (interface{}, string, error) {
		vm, err := c.GetVm(Vm{Id: id})

		if err != nil {
			return vm, "", err
		}

		return vm, vm.PowerState, nil
	}
}

func (c *Client) waitForPVDriversDetected(id string) error {
	refreshFn := func() (result interface{}, state string, err error) {
		vm, err := c.GetVm(Vm{Id: id})

		if err != nil {
			return vm, "", err
		}

		return vm, strconv.FormatBool(vm.PVDriversDetected), nil
	}
	stateConf := &StateChangeConf{
		Pending: []string{"false"},
		Refresh: refreshFn,
		Target:  []string{"true"},
		Timeout: 2 * time.Minute,
	}
	_, err := stateConf.WaitForState()
	return err
}

func (c *Client) waitForVmState(id string, stateConf StateChangeConf) error {
	stateConf.Refresh = GetVmPowerState(c, id)
	_, err := stateConf.WaitForState()
	return err
}

func waitForPowerStateReached(c *Client, vmId, desiredPowerState string, timeout time.Duration) error {
	var pending []string
	target := desiredPowerState
	switch desiredPowerState {
	case RunningPowerState:
		pending = []string{HaltedPowerState}
	case HaltedPowerState:
		pending = []string{RunningPowerState}
	default:
		return errors.New(fmt.Sprintf("Invalid VM power state requested: %s\n", desiredPowerState))
	}
	refreshFn := func() (result interface{}, state string, err error) {
		vm, err := c.GetVm(Vm{Id: vmId})

		if err != nil {
			return vm, "", err
		}

		return vm, vm.PowerState, nil
	}
	stateConf := &StateChangeConf{
		Pending: pending,
		Refresh: refreshFn,
		Target:  []string{target},
		Timeout: timeout,
	}
	_, err := stateConf.WaitForState()
	return err
}

type ifaceMatchCheck struct {
	cidrRange  string
	ifaceIdx   string
	ifaceAddrs []string
}

func waitForIPAssignment(c *Client, vmId string, waitForIps map[string]string, timeout time.Duration) error {
	var lastResult ifaceMatchCheck
	refreshFn := func() (result interface{}, state string, err error) {
		vm, err := c.GetVm(Vm{Id: vmId})

		if err != nil {
			return vm, "", err
		}

		addrs := vm.Addresses
		if len(addrs) == 0 || vm.PowerState != RunningPowerState {
			return addrs, "Waiting", nil
		}

		netIfaces := map[string][]string{}
		for key, addr := range vm.Addresses {

			// key has the following format "{iface_id}/(ipv4|ipv6)/{iface_ip_id}"
			ifaceIdx, _, _ := strings.Cut(key, "/")
			if _, ok := netIfaces[ifaceIdx]; !ok {
				netIfaces[ifaceIdx] = []string{}
			}
			netIfaces[ifaceIdx] = append(netIfaces[ifaceIdx], addr)
		}

		for ifaceIdx, cidrRange := range waitForIps {
			// VM's Addresses member does not contain this network interface yet
			if _, ok := netIfaces[ifaceIdx]; !ok {
				return addrs, "Waiting", nil
			}

			found := false
			for _, ipAddr := range netIfaces[ifaceIdx] {
				_, ipNet, err := net.ParseCIDR(cidrRange)

				if err != nil {
					return addrs, "Waiting", err
				}

				if ipNet.Contains(net.ParseIP(ipAddr)) {
					found = true
				}
			}

			if !found {
				lastResult = ifaceMatchCheck{
					cidrRange:  cidrRange,
					ifaceIdx:   ifaceIdx,
					ifaceAddrs: netIfaces[ifaceIdx],
				}

				return addrs, "Waiting", nil
			}
		}

		return addrs, "Ready", nil
	}
	stateConf := &StateChangeConf{
		Pending: []string{"Waiting"},
		Refresh: refreshFn,
		Target:  []string{"Ready"},
		Timeout: timeout,
	}
	_, err := stateConf.WaitForState()
	if _, ok := err.(*TimeoutError); ok {
		return errors.New(fmt.Sprintf("network[%s] never converged to the following cidr: %s, addresses: %s failed to match", lastResult.ifaceIdx, lastResult.cidrRange, lastResult.ifaceAddrs))
	}
	return err
}

func (c *Client) waitForModifyVm(id string, desiredPowerState string, waitForIps map[string]string, timeout time.Duration) error {
	if len(waitForIps) == 0 {
		return waitForPowerStateReached(c, id, desiredPowerState, timeout)
	} else {
		return waitForIPAssignment(c, id, waitForIps, timeout)
	}
}

func FindOrCreateVmForTests(vm *Vm, poolId, srId, templateName, tag string) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v\n", err)
		os.Exit(-1)
	}

	var vmRes *Vm
	var net *Network
	vmRes, err = c.GetVm(Vm{
		Tags: []string{tag},
	})

	if _, ok := err.(NotFound); ok {
		net, err = c.GetNetwork(Network{
			// We assume that a eth0 pool wide network exists
			// since trying to discern what the appropriate network
			// is from our current set of test inputs is challenging.
			// If this proves problematic then it can be reconsidered.
			NameLabel: "Pool-wide network associated with eth0",
			PoolId:    poolId,
		})

		if err != nil {
			fmt.Println("Failed to get network to create a Vm for the tests")
			os.Exit(-1)
		}

		vmRes, err = c.CreateVm(
			Vm{
				NameLabel: fmt.Sprintf("Terraform testing - %d", time.Now().Unix()),
				Tags:      []string{tag},
				Template:  templateName,
				CPUs: CPUs{
					Number: 1,
				},
				Memory: MemoryObject{
					Static: []int{
						0, 2147483648,
					},
				},
				VIFsMap: []map[string]string{
					{
						"network": net.Id,
					},
				},
				Disks: []Disk{
					Disk{
						VDI: VDI{
							SrId:      srId,
							NameLabel: "terraform xenorchestra client testing",
							Size:      16106127360,
						},
					},
					Disk{
						VDI: VDI{
							SrId:      srId,
							NameLabel: "disk2",
							Size:      16106127360,
						},
					},
				},
			},
			5*time.Minute,
		)
	}

	if err != nil {
		fmt.Println(fmt.Sprintf("failed to find vm for the client tests with error: %v\n", err))
		os.Exit(-1)
	}

	*vm = *vmRes
}

func checkBlockDestroyOperation(vm *Vm) bool {
	fmt.Printf("Found VM with blocked_operations=%v", vm.BlockedOperations)

	for k, _ := range vm.BlockedOperations {

		if k == "destroy" {
			return true
		}

	}
	return false
}

func RemoveVmsWithNamePrefix(prefix string) func(string) error {
	return func(_ string) error {
		fmt.Println("[DEBUG] Running vm sweeper")
		c, err := NewClient(GetConfigFromEnv())
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		var vmsMap map[string]Vm
		err = c.GetAllObjectsOfType(Vm{}, &vmsMap)
		if err != nil {
			return fmt.Errorf("error getting vms: %s", err)
		}
		for _, vm := range vmsMap {
			if strings.HasPrefix(vm.NameLabel, prefix) {
				if checkBlockDestroyOperation(&vm) {
					var success bool
					blockedOperations := map[string]interface{}{
						"destroy": nil,
					}
					params := map[string]interface{}{
						"id":                vm.Id,
						"blockedOperations": blockedOperations,
					}
					client, _ := c.(*Client)
					err := client.Call("vm.set", params, &success)

					if err != nil {
						log.Printf("error removing destroy block on vm `%s` during sweep: %s", vm.NameLabel, err)
					}
				}
				fmt.Printf("[DEBUG] Deleting vm `%s`\n", vm.NameLabel)
				err := c.DeleteVm(vm.Id)
				if err != nil {
					log.Printf("error destroying vm `%s` during sweep: %s", vm.NameLabel, err)
				}
			}
		}
		return nil
	}
}

// This is not meant to be a robust check since it would be complicated to detect all
// malformed config. The goal is to cover what is supported by the cloudinit terraform
// provider (https://github.com/hashicorp/terraform-provider-cloudinit) and to rule out
// obviously bad config
func warnOnInvalidCloudConfig(cloudConfig string) {
	contentType := http.DetectContentType([]byte(cloudConfig))
	if contentType == "application/x-gzip" {
		return
	}

	if strings.HasPrefix(cloudConfig, "Content-Type") {
		if !strings.Contains(cloudConfig, "multipart/") {

			log.Printf("[WARNING] Detected MIME type that may not be supported by cloudinit")
			log.Printf("[WARNING] Validate that your configuration is well formed according to the documentation (https://cloudinit.readthedocs.io/en/latest/topics/format.html).\n")
		}
		return
	}
	if !strings.HasPrefix(cloudConfig, "#cloud-config") {
		log.Printf("[WARNING] cloud config does not start with required text `#cloud-config`.")
		log.Printf("[WARNING] Validate that your configuration is well formed according to the documentation (https://cloudinit.readthedocs.io/en/latest/topics/format.html).\n")
	}

}
