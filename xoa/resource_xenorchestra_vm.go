package xoa

import (
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa/internal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validVga = []string{
	"",
	"cirrus",
	"std",
}

var validHaOptions = []string{
	"",
	"best-effort",
	"restart",
}

var validFirmware = []string{
	"bios",
	"uefi",
}

var validInstallationMethods = []string{
	"network",
}

func resourceVm() *schema.Resource {
	vmSchema := resourceVmSchema()
	delete(vmSchema, "cdrom")
	delete(vmSchema, "installation_method")
	vmSchema["id"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	}

	return &schema.Resource{
		Schema: vmSchema,
	}
}

func resourceVmSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{

		"affinity_host": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"blocked_operations": &schema.Schema{
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"name_label": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		"name_description": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"cloud_network_config": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"auto_poweron": &schema.Schema{
			Type:     schema.TypeBool,
			Default:  false,
			Optional: true,
		},
		"exp_nested_hvm": &schema.Schema{
			Type:     schema.TypeBool,
			Default:  false,
			Optional: true,
		},
		"hvm_boot_firmware": &schema.Schema{
			Type:         schema.TypeString,
			Default:      "bios",
			Optional:     true,
			ValidateFunc: validation.StringInSlice(validFirmware, false),
		},
		"power_state": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
		"installation_method": &schema.Schema{
			Type:          schema.TypeString,
			Optional:      true,
			ValidateFunc:  validation.StringInSlice(validInstallationMethods, false),
			ConflictsWith: []string{"cdrom"},
		},
		"high_availability": &schema.Schema{
			Type:         schema.TypeString,
			Default:      "",
			Optional:     true,
			ValidateFunc: validation.StringInSlice(validHaOptions, false),
		},
		"template": &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"cloud_config": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"core_os": &schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"cpu_cap": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"cpu_weight": &schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Default:  0,
		},
		"cpus": &schema.Schema{
			Type:     schema.TypeInt,
			Required: true,
		},
		"memory_max": &schema.Schema{
			Type:     schema.TypeInt,
			Required: true,
		},
		"resource_set": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"ipv4_addresses": &schema.Schema{
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"ipv6_addresses": &schema.Schema{
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"vga": &schema.Schema{
			Type:         schema.TypeString,
			Default:      "std",
			Optional:     true,
			ValidateFunc: validation.StringInSlice(validVga, false),
		},
		"videoram": &schema.Schema{
			Type:     schema.TypeInt,
			Default:  8,
			Optional: true,
		},
		"start_delay": &schema.Schema{
			Type:     schema.TypeInt,
			Default:  0,
			Optional: true,
		},
		// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
		// "secure_boot": &schema.Schema{
		// 	Type:     schema.TypeBool,
		// 	Default:  false,
		// 	Optional: true,
		// },
		"host": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"wait_for_ip": &schema.Schema{
			Type:     schema.TypeBool,
			Default:  false,
			Optional: true,
		},
		"cdrom": &schema.Schema{
			Type:          schema.TypeList,
			Optional:      true,
			ConflictsWith: []string{"installation_method"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			// This should be increased but I don't understand
			// the use cases for multiple ISOs just yet. For now
			// limit it to a single ISO
			MaxItems: 1,
		},
		"network": &schema.Schema{
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"attached": &schema.Schema{
						Type:             schema.TypeBool,
						Default:          true,
						Optional:         true,
						DiffSuppressFunc: suppressAttachedDiffWhenHalted,
					},
					"device": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
					"network_id": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"mac_address": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						StateFunc: func(val interface{}) string {
							unformattedMac := val.(string)
							mac, err := net.ParseMAC(unformattedMac)
							if err != nil {
								panic(fmt.Sprintf("Mac address `%s` was not parsable. This should never happened because Terraform's validation should happen before this is stored into state", unformattedMac))
							}
							return mac.String()

						},
						ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
							mac_address := val.(string)
							if _, err := net.ParseMAC(mac_address); err != nil {
								errs = append(errs, fmt.Errorf("%s Mac Address is invalid", mac_address))
							}
							return

						},
					},
					"ipv4_addresses": &schema.Schema{
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"ipv6_addresses": &schema.Schema{
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"disk": &schema.Schema{
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"sr_id": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"name_label": &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					},
					"name_description": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
					},
					"size": &schema.Schema{
						Type:     schema.TypeInt,
						Required: true,
					},
					"attached": &schema.Schema{
						Type:             schema.TypeBool,
						Default:          true,
						Optional:         true,
						DiffSuppressFunc: suppressAttachedDiffWhenHalted,
					},
					"position": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
					"vdi_id": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
					"vbd_id": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"tags": resourceTags(),
	}
}

func resourceRecord() *schema.Resource {
	duration := 5 * time.Minute
	return &schema.Resource{
		Create: resourceVmCreate,
		Read:   resourceVmRead,
		Update: resourceVmUpdate,
		Delete: resourceVmDelete,
		Importer: &schema.ResourceImporter{
			State: RecordImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: &duration,
			Update: &duration,
			Delete: &duration,
		},
		Schema: resourceVmSchema(),
	}
}

func resourceVmCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	network_maps := []map[string]string{}
	networks := d.Get("network").([]interface{})

	for _, network := range networks {
		netMap, _ := network.(map[string]interface{})

		netID := netMap["network_id"].(string)
		macAddr := netMap["mac_address"].(string)

		netMapToAdd := map[string]string{
			"network": netID,
		}
		// We only add the mac address if it contains a value.
		if macAddr != "" {
			netMapToAdd["mac"] = getFormattedMac(macAddr)
		}

		network_maps = append(network_maps, netMapToAdd)
	}

	ds := []client.Disk{}

	disks := d.Get("disk").([]interface{})

	for _, disk := range disks {
		vdi, _ := disk.(map[string]interface{})

		ds = append(ds, client.Disk{
			VDI: client.VDI{
				SrId:            vdi["sr_id"].(string),
				NameLabel:       vdi["name_label"].(string),
				NameDescription: vdi["name_description"].(string),
				Size:            vdi["size"].(int),
			},
		})
	}

	tags := d.Get("tags").(*schema.Set).List()
	vmTags := []string{}
	for _, tag := range tags {
		t := tag.(string)
		vmTags = append(vmTags, t)
	}

	cds := []string{}
	for _, cd := range d.Get("cdrom").([]interface{}) {
		cdMap := cd.(map[string]interface{})
		cds = append(cds, cdMap["id"].(string))
	}
	installation := client.Installation{}
	if len(cds) > 0 {
		installation = client.Installation{
			Method:     "cdrom",
			Repository: cds[0],
		}
	}

	if installMethod := d.Get("installation_method").(string); installMethod != "" {
		installation = client.Installation{
			Method: "network",
		}
	}

	blockedOperations := map[string]string{}
	for _, b := range d.Get("blocked_operations").(*schema.Set).List() {
		blockedOperations[b.(string)] = "true"
	}

	var rs *client.FlatResourceSet
	if rsId, ok := d.GetOk("resource_set"); ok {
		rs = &client.FlatResourceSet{
			Id: rsId.(string),
		}
	}
	vm, err := c.CreateVm(client.Vm{
		AffinityHost:      d.Get("affinity_host").(string),
		BlockedOperations: blockedOperations,
		Boot: client.Boot{
			Firmware: d.Get("hvm_boot_firmware").(string),
		},
		ExpNestedHvm:    d.Get("exp_nested_hvm").(bool),
		NameLabel:       d.Get("name_label").(string),
		NameDescription: d.Get("name_description").(string),
		Template:        d.Get("template").(string),
		CloudConfig:     d.Get("cloud_config").(string),
		ResourceSet:     rs,
		CPUs: client.CPUs{
			Number: d.Get("cpus").(int),
		},
		CloudNetworkConfig: d.Get("cloud_network_config").(string),
		Memory: client.MemoryObject{
			Static: []int{
				0, d.Get("memory_max").(int),
			},
		},
		Tags:         vmTags,
		Disks:        ds,
		Installation: installation,
		// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
		// SecureBoot:   d.Get("secure_boot").(bool),
		VIFsMap:    network_maps,
		StartDelay: d.Get("start_delay").(int),
		WaitForIps: d.Get("wait_for_ip").(bool),
		Videoram: client.Videoram{
			Value: d.Get("videoram").(int),
		},
		Vga: d.Get("vga").(string),
	},
		d.Timeout(schema.TimeoutCreate),
	)

	if err != nil {
		return err
	}

	vifs, err := c.GetVIFs(vm)

	if err != nil {
		return err
	}

	vmDisks, err := c.GetDisks(vm)

	if err != nil {
		return err
	}

	cdroms, err := c.GetCdroms(vm)
	if err != nil {
		return err
	}

	return recordToData(*vm, vifs, vmDisks, cdroms, d)
}

func sortDiskByPostion(disks []client.Disk) []client.Disk {
	sort.Slice(disks, func(i, j int) bool {
		one, _ := strconv.Atoi(disks[i].Position)
		other, _ := strconv.Atoi(disks[j].Position)
		return one < other
	})
	return disks
}

func sortNetworksByDevice(networks []*client.VIF) []*client.VIF {
	sort.Slice(networks, func(i, j int) bool {
		one, _ := strconv.Atoi(networks[i].Device)
		other, _ := strconv.Atoi(networks[j].Device)
		return one < other
	})
	return networks
}

func sortNetworkMapByDevice(networks []map[string]interface{}) []map[string]interface{} {

	sort.Slice(networks, func(i, j int) bool {
		one, _ := strconv.Atoi(networks[i]["device"].(string))
		other, _ := strconv.Atoi(networks[j]["device"].(string))
		return one < other
	})
	return networks
}

func sortDiskMapByPostion(networks []map[string]interface{}) []map[string]interface{} {

	sort.Slice(networks, func(i, j int) bool {
		one, _ := strconv.Atoi(networks[i]["position"].(string))
		other, _ := strconv.Atoi(networks[j]["position"].(string))
		return one < other
	})
	return networks
}

func disksToMapList(disks []client.Disk) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(disks))
	for _, disk := range disks {
		if disk.NameLabel == "XO CloudConfigDrive" {
			continue
		}
		diskMap := map[string]interface{}{
			"attached":         disk.Attached,
			"vbd_id":           disk.Id,
			"vdi_id":           disk.VDIId,
			"position":         disk.Position,
			"name_label":       disk.NameLabel,
			"name_description": disk.NameDescription,
			"size":             disk.Size,
			"sr_id":            disk.SrId,
		}
		result = append(result, diskMap)
	}

	return sortDiskMapByPostion(result)
}

func cdromsToMapList(disks []client.Disk) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(disks))
	for _, disk := range disks {
		diskMap := map[string]interface{}{
			"id": disk.VDIId,
		}
		result = append(result, diskMap)
	}

	return result
}

func vifsToMapList(vifs []client.VIF, guestNets []guestNetwork) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(vifs))
	for _, vif := range vifs {
		ipv6Addrs := []string{}
		ipv4Addrs := []string{}
		device, _ := strconv.Atoi(vif.Device)
		log.Printf("[DEBUG] Trying to find ip addresses for device '%d' in guest networks: %v\n", device, guestNets)
		if len(guestNets) > device {
			ipv4Addrs = guestNets[device]["ipv4"]
			ipv6Addrs = guestNets[device]["ipv6"]
		}
		vifMap := map[string]interface{}{
			"attached":       vif.Attached,
			"device":         vif.Device,
			"mac_address":    vif.MacAddress,
			"network_id":     vif.Network,
			"ipv4_addresses": ipv4Addrs,
			"ipv6_addresses": ipv6Addrs,
		}
		result = append(result, vifMap)
	}

	return sortNetworkMapByDevice(result)
}

func resourceVmRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	vm, err := c.GetVm(client.Vm{Id: d.Id()})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	vifs, err := c.GetVIFs(vm)

	if err != nil {
		return err
	}

	disks, err := c.GetDisks(vm)

	if err != nil {
		return err
	}

	cdroms, err := c.GetCdroms(vm)
	if err != nil {
		return err
	}

	return recordToData(*vm, vifs, disks, cdroms, d)
}

func resourceVmUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	id := d.Id()
	nameLabel := d.Get("name_label").(string)
	affinityHost := d.Get("affinity_host").(string)
	nameDescription := d.Get("name_description").(string)
	cpus := d.Get("cpus").(int)
	autoPowerOn := d.Get("auto_poweron").(bool)
	ha := d.Get("high_availability").(string)
	var rs *client.FlatResourceSet
	if d.HasChange("resource_set") {
		rs = &client.FlatResourceSet{
			Id: d.Get("resource_set").(string),
		}
	}
	memoryMax := d.Get("memory_max").(int)

	vm, err := c.GetVm(client.Vm{Id: id})

	if err != nil {
		return err
	}

	if d.HasChange("network") {
		origNet, newNet := d.GetChange("network")
		oSet := schema.NewSet(vifHash, origNet.([]interface{}))
		nSet := schema.NewSet(vifHash, newNet.([]interface{}))

		removals := expandNetworks(oSet.Difference(nSet).List())
		log.Printf("[DEBUG] Found the following network removals: %v previous set: %v new set: %v\n", oSet.Difference(nSet).List(), oSet, nSet)
		for _, removal := range removals {
			// We will process the updates with the additons so we only need to deal with
			// VIFs that need to be removed.
			updateVif, _ := shouldUpdateVif(*removal, expandNetworks(nSet.List()))
			if updateVif {
				continue
			} else {

				vifErr := c.DeleteVIF(removal)

				if vifErr != nil {
					return vifErr
				}
			}
		}

		additions := sortNetworksByDevice(expandNetworks(nSet.Difference(oSet).List()))
		log.Printf("[DEBUG] Found the following network additions: %v previous set: %v new set: %v\n", nSet.Difference(oSet).List(), oSet, nSet)
		for _, addition := range additions {
			updateVif, shouldAttach := shouldUpdateVif(*addition, expandNetworks(oSet.List()))
			var vifErr error
			if updateVif {
				switch shouldAttach {
				case true:
					vifErr = c.ConnectVIF(addition)
				case false:
					vifErr = c.DisconnectVIF(addition)
				}
				if vifErr != nil {
					return vifErr
				}
			} else {
				_, vifErr := c.CreateVIF(vm, addition)

				if vifErr != nil {
					return vifErr
				}
			}
		}
	}

	if d.HasChange("cdrom") {
		oCds, nCds := d.GetChange("cdrom")

		for range oCds.([]interface{}) {
			err := c.EjectCd(id)

			if err != nil {
				return err
			}
		}

		for _, cd := range nCds.([]interface{}) {
			cdMap := cd.(map[string]interface{})
			cdId := cdMap["id"].(string)
			err := c.InsertCd(id, cdId)

			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("disk") {
		oDisk, nDisk := d.GetChange("disk")

		oSet := schema.NewSet(diskHash, oDisk.([]interface{}))
		nSet := schema.NewSet(diskHash, nDisk.([]interface{}))

		removals := expandDisks(oSet.Difference(nSet).List())
		log.Printf("[DEBUG] Found the following disk removals: %v previous set: %v new set: %v\n", oSet.Difference(nSet).List(), oSet, nSet)
		for _, removal := range removals {

			actions := getUpdateDiskActions(removal, expandDisks(nSet.List()))
			if len(actions) != 0 {
				continue
			}

			if err := c.DeleteDisk(*vm, removal); err != nil {
				return err
			}
		}

		additions := sortDiskByPostion(expandDisks(nSet.Difference(oSet).List()))
		log.Printf("[DEBUG] Found the following disk additions: %v previous set: %v new set: %v\n", nSet.Difference(oSet).List(), oSet, nSet)
		for _, disk := range additions {

			actions := getUpdateDiskActions(disk, expandDisks(oSet.List()))
			log.Printf("[DEBUG] Found '%v' disk update actions\n", actions)

			if len(actions) == 0 {
				if _, err := c.CreateDisk(*vm, disk); err != nil {
					return err
				}
				continue
			}

			for _, action := range actions {
				log.Printf("[DEBUG] Updating disk with action '%d'\n", action)
				if err := performDiskUpdateAction(c, action, disk); err != nil {
					return err
				}
			}
		}
	}

	haltForUpdates := false
	if _, nCPUs := d.GetChange("cpus"); d.HasChange("cpus") && nCPUs.(int) > vm.CPUs.Max {
		haltForUpdates = true
	}

	if d.HasChange("memory_max") {
		haltForUpdates = true
	}

	blockOperations := map[string]string{}
	if d.HasChange("blocked_operations") {
		o, n := d.GetChange("blocked_operations")
		oldBlockedOps := o.(*schema.Set)
		newBlockedOps := n.(*schema.Set)

		oB := oldBlockedOps.Difference(newBlockedOps)
		for _, removal := range oB.List() {
			blockOperations[removal.(string)] = "false"
		}

		nB := newBlockedOps.Difference(oldBlockedOps)
		for _, addition := range nB.List() {
			blockOperations[addition.(string)] = "true"
		}

	}

	vmReq := client.Vm{
		Id: id,
		CPUs: client.CPUs{
			Number: cpus,
		},
		Memory: client.MemoryObject{
			Static: []int{
				0, memoryMax,
			},
		},
		NameLabel:         nameLabel,
		NameDescription:   nameDescription,
		HA:                ha,
		ResourceSet:       rs,
		AutoPoweron:       autoPowerOn,
		AffinityHost:      affinityHost,
		BlockedOperations: blockOperations,
		ExpNestedHvm:      d.Get("exp_nested_hvm").(bool),
		StartDelay:        d.Get("start_delay").(int),
		Vga:               d.Get("vga").(string),
		// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
		// SecureBoot:        d.Get("secure_boot").(bool),
		Boot: client.Boot{
			Firmware: d.Get("hvm_boot_firmware").(string),
		},
		Videoram: client.Videoram{
			Value: d.Get("videoram").(int),
		},
	}
	if haltForUpdates {
		err := c.HaltVm(id)

		if err != nil {
			return err
		}
	}
	vm, err = c.UpdateVm(vmReq)

	if haltForUpdates {
		err := c.StartVm(vmReq.Id)

		if err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Retrieved vm after update: %+v\n", vm)

	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		oTags := o.(*schema.Set)
		nTags := n.(*schema.Set)

		removals := oTags.Difference(nTags)
		for _, removal := range removals.List() {
			if err := c.RemoveTag(id, removal.(string)); err != nil {
				return err
			}
		}

		additions := nTags.Difference(oTags)
		for _, addition := range additions.List() {
			if err := c.AddTag(id, addition.(string)); err != nil {
				return err
			}
		}
	}

	return resourceVmRead(d, m)
}

func resourceVmDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	err := c.DeleteVm(d.Id())

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func expandDisks(disks []interface{}) []client.Disk {
	result := make([]client.Disk, 0, len(disks))

	for _, disk := range disks {
		data := disk.(map[string]interface{})

		result = append(result, client.Disk{
			client.VBD{
				Id:       data["vbd_id"].(string),
				Attached: data["attached"].(bool),
			},
			client.VDI{
				VDIId:           data["vdi_id"].(string),
				NameLabel:       data["name_label"].(string),
				NameDescription: data["name_description"].(string),
				SrId:            data["sr_id"].(string),
				Size:            data["size"].(int),
			},
		})
	}

	return result
}

func expandNetworks(networks []interface{}) []*client.VIF {
	vifs := make([]*client.VIF, 0, len(networks))
	for _, net := range networks {
		data := net.(map[string]interface{})

		attached := data["attached"].(bool)
		device := data["device"].(string)
		networkId := data["network_id"].(string)
		macAddress := getFormattedMac(data["mac_address"].(string))
		vifs = append(vifs, &client.VIF{
			Attached:   attached,
			Device:     device,
			Network:    networkId,
			MacAddress: macAddress,
		})
	}
	return vifs
}

func RecordImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(client.XOClient)

	vm, err := c.GetVm(client.Vm{Id: d.Id()})
	if err != nil {
		return nil, err
	}

	rd := []*schema.ResourceData{d}
	vifs, err := c.GetVIFs(vm)

	if err != nil {
		return rd, err
	}

	disks, err := c.GetDisks(vm)

	if err != nil {
		return rd, err
	}

	cdroms, err := c.GetCdroms(vm)
	if err != nil {
		return rd, err
	}

	err = recordToData(*vm, vifs, disks, cdroms, d)

	return rd, err
}

func recordToData(resource client.Vm, vifs []client.VIF, disks []client.Disk, cdroms []client.Disk, d *schema.ResourceData) error {
	d.SetId(resource.Id)
	// d.Set("cloud_config", resource.CloudConfig)
	if len(resource.Memory.Dynamic) == 2 {
		if err := d.Set("memory_max", resource.Memory.Dynamic[1]); err != nil {
			return err
		}
	} else {
		log.Printf("[WARN] Expected the VM's static memory limits to have two values, %v found instead\n", resource.Memory.Dynamic)
	}

	d.Set("cpus", resource.CPUs.Number)
	d.Set("name_label", resource.NameLabel)
	d.Set("affinity_host", resource.AffinityHost)
	d.Set("name_description", resource.NameDescription)
	d.Set("high_availability", resource.HA)
	d.Set("auto_poweron", resource.AutoPoweron)
	if resource.ResourceSet != nil {
		d.Set("resource_set", resource.ResourceSet.Id)
	} else {
		d.Set("resource_set", "")
	}
	d.Set("power_state", resource.PowerState)

	// TODO: (#145) Uncomment this once issues with secure_boot have been figured out
	// if err := d.Set("secure_boot", resource.SecureBoot); err != nil {
	// 	return err
	// }

	if err := d.Set("hvm_boot_firmware", resource.Boot.Firmware); err != nil {
		return err
	}

	if err := d.Set("exp_nested_hvm", resource.ExpNestedHvm); err != nil {
		return err
	}

	if err := d.Set("vga", resource.Vga); err != nil {
		return err
	}

	if err := d.Set("videoram", resource.Videoram.Value); err != nil {
		return err
	}

	if err := d.Set("start_delay", resource.StartDelay); err != nil {
		return err
	}

	if err := d.Set("tags", resource.Tags); err != nil {
		return err
	}
	if err := d.Set("blocked_operations", vmBlockedOperationsToList(resource)); err != nil {
		return err
	}

	log.Printf("[DEBUG] Found the following ip addresses: %v\n", resource.Addresses)
	networkIps := extractIpsFromNetworks(resource.Addresses)
	nets := vifsToMapList(vifs, networkIps)
	err := d.Set("network", nets)

	if err != nil {
		return err
	}

	disksMapList := disksToMapList(disks)
	err = d.Set("disk", disksMapList)
	if err != nil {
		return err
	}

	cdsMapList := cdromsToMapList(cdroms)
	err = d.Set("cdrom", cdsMapList)
	if err != nil {
		return err
	}

	if len(networkIps) > 0 {
		for _, proto := range []string{"ipv4", "ipv6"} {
			ips := []string{}
			for i := range networkIps {
				ips = append(ips, networkIps[i][proto]...)
				if err := d.Set(fmt.Sprintf("%s_addresses", proto), ips); err != nil {
					return err
				}
			}

		}
	} else {
		if err := d.Set("ipv4_addresses", []string{}); err != nil {
			return err
		}
		if err := d.Set("ipv6_addresses", []string{}); err != nil {
			return err
		}
	}

	return nil
}

func vmBlockedOperationsToList(v client.Vm) []string {
	blockedOperations := []string{}
	for k, _ := range v.BlockedOperations {
		blockedOperations = append(blockedOperations, k)
	}

	return blockedOperations
}

func diskHash(value interface{}) int {
	var srId string
	var nameLabel string
	var nameDescription string
	var size int
	var attached bool
	switch t := value.(type) {
	case client.Disk:
		srId = t.SrId
		nameLabel = t.NameLabel
		nameDescription = t.NameDescription
		size = t.Size
		attached = t.Attached
	case map[string]interface{}:
		srId = t["sr_id"].(string)
		nameLabel = t["name_label"].(string)
		nameDescription = t["name_description"].(string)
		size = t["size"].(int)
		attached = t["attached"].(bool)
	default:
		panic(fmt.Sprintf("disk cannot be hashed with type %T", t))
	}
	v := fmt.Sprintf("%s-%s-%s-%d-%t", nameLabel, nameDescription, srId, size, attached)
	return internal.String(v)
}

func vifHash(value interface{}) int {
	var macAddress string
	var networkId string
	var device string
	var attached bool
	switch t := value.(type) {
	case client.VIF:
		macAddress = t.MacAddress
		networkId = t.Network
		device = t.Device
		attached = t.Attached
	case map[string]interface{}:
		network := value.(map[string]interface{})
		macAddress = network["mac_address"].(string)
		networkId = network["network_id"].(string)
		device = network["device"].(string)
		attached = network["attached"].(bool)
	default:
		panic(fmt.Sprintf("can't has type %T", t))
	}

	v := fmt.Sprintf("%s-%s-%s-%t", macAddress, networkId, device, attached)
	log.Printf("[TRACE] Using the following as input to the VIF hash function: %s\n", v)

	return internal.String(v)
}

func shouldUpdateVif(vif client.VIF, vifs []*client.VIF) (bool, bool) {
	found := false
	vifCopy := vif
	var vifFound client.VIF
	for _, vifToCheck := range vifs {
		if vifToCheck.Id == vifCopy.Id || vifToCheck.MacAddress == vifCopy.MacAddress {
			found = true
			vifFound = *vifToCheck
		}
	}

	vifFound.Attached = !vifFound.Attached
	if found && vifHash(vifCopy) == vifHash(vifFound) {
		return true, vifCopy.Attached
	}

	return false, false
}

type updateDiskActions int

const (
	diskNameDescriptionUpdate updateDiskActions = iota
	diskNameLabelUpdate
	diskAttachmentUpdate
)

func getUpdateDiskActions(d client.Disk, disks []client.Disk) []updateDiskActions {
	actions := []updateDiskActions{}
	var diskFound *client.Disk
	for _, disk := range disks {
		if disk.Id == d.Id {
			diskFound = &disk
		}
	}

	if diskFound == nil {
		return actions
	}

	if diskFound.NameLabel != d.NameLabel {
		actions = append(actions, diskNameLabelUpdate)
	}

	if diskFound.NameDescription != d.NameDescription {
		actions = append(actions, diskNameDescriptionUpdate)
	}

	if diskFound.Attached != d.Attached {
		actions = append(actions, diskAttachmentUpdate)
	}
	return actions
}

func shouldUpdateDisk(d client.Disk, disks []client.Disk) bool {
	found := false
	var diskFound client.Disk
	for _, disk := range disks {
		if disk.Id == d.Id {
			found = true
			diskFound = disk
		}
	}

	diskFound.Attached = !diskFound.Attached
	if found && diskHash(diskFound) == diskHash(d) {
		return true
	}
	return false
}

func performDiskUpdateAction(c client.XOClient, action updateDiskActions, d client.Disk) error {
	switch action {
	case diskAttachmentUpdate:
		if d.Attached {
			return c.ConnectDisk(d)
		} else {
			return c.DisconnectDisk(d)
		}
	case diskNameDescriptionUpdate:
		return c.UpdateVDI(d)
	case diskNameLabelUpdate:
		return c.UpdateVDI(d)
	}

	return errors.New(fmt.Sprintf("disk update action '%d' not handled", action))
}

func getFormattedMac(macAddress string) string {
	if macAddress == "" {
		return macAddress
	}
	mac, err := net.ParseMAC(macAddress)

	if err != nil {
		panic(fmt.Sprintf("Mac address `%s` was not parsable. This is a bug in the provider and this value should have been properly formatted", macAddress))
	}
	return mac.String()
}

type guestNetwork map[string][]string

// Transforms Xen's guest-tools network information into a slice
// of maps where each element represents a network interface.
// Each map will contain the following keys: ip, ipv4 and ipv6. The values
// will be a slice of ip addresses.
//
//	[]map[string][]string{
//	  {
//	    "ip":   []string{"interface 1's IPs",
//	    "ipv4": []string{"interface 1's IPs",
//	    "ipv6": []string{"ip1", "ip2"}
//	  },
//	  {
//	    "ip":   []string{"interface 2's IPs",
//	    "ipv4": []string{"interface 2's IPs",
//	    "ipv6": []string{"ip1", "ip2"}
//	  },
//	}
func extractIpsFromNetworks(networks map[string]string) []guestNetwork {

	if len(networks) < 1 {
		return []guestNetwork{}
	}

	IP_REGEX := `^(\d+)\/(ip(?:v4|v6)?)(?:\/(\d+))?$`
	reg := regexp.MustCompile(IP_REGEX)

	keys := make([]string, len(networks))
	for k := range networks {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	last := keys[len(keys)-1]

	matches := reg.FindStringSubmatch(last)
	if matches == nil || len(matches) != 4 {
		panic("this should never happen")
	}
	cap, _ := strconv.Atoi(matches[1])
	devices := make([]guestNetwork, 0, cap)
	for _, key := range keys {
		matches := reg.FindStringSubmatch(key)

		if matches == nil || len(matches) != 4 {
			continue
		}
		deviceStr, proto := matches[1], matches[2]

		deviceNum, _ := strconv.Atoi(deviceStr)
		for len(devices) <= deviceNum {
			devices = append(devices, map[string][]string{})
		}

		// This will panic. Need to use for loop like above
		ipList := devices[deviceNum][proto]
		if ipList == nil {
			devices[deviceNum][proto] = []string{}
		}

		devices[deviceNum][proto] = append(devices[deviceNum][proto], networks[key])
	}
	log.Printf("[DEBUG] Extracted the following network interface ips: %v\n", devices)
	return devices
}

func suppressAttachedDiffWhenHalted(k, old, new string, d *schema.ResourceData) (suppress bool) {
	powerState := d.Get("power_state").(string)
	suppress = true

	if powerState == "Running" {
		suppress = false
	}
	log.Printf("[DEBUG] VM '%s' attribute has transitioned from '%s' to '%s' when PowerState '%s'. Suppress diff: %t", k, old, new, powerState, suppress)
	return
}
