package xoa

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func init() {
}

var validHaOptions = []string{
	"",
	"best-effort",
	"restart",
}

func StringInSlice(valid []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		for _, str := range valid {
			if v == str || (ignoreCase && strings.ToLower(v) == strings.ToLower(str)) {
				return
			}
		}

		es = append(es, fmt.Errorf("expected %s to be one of %v, got %s", k, valid, v))
		return
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
		},
		Schema: map[string]*schema.Schema{
			"name_label": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name_description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_poweron": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"high_availability": &schema.Schema{
				Type:     schema.TypeString,
				Default:  "",
				Optional: true,
				// TODO: Replace with validation.StringInSlice when terraform
				// and the SDK are upgraded.
				ValidateFunc: StringInSlice(validHaOptions, false),
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
			"network": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attached": &schema.Schema{
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
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
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								mac_address := val.(string)
								if _, err := net.ParseMAC(mac_address); err != nil {
									errs = append(errs, fmt.Errorf("%s Mac Address is invalid", mac_address))
								}
								return

							},
						},
					},
				},
			},
			"disk": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// 	return false
				// },
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
						"size": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"attached": &schema.Schema{
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
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
		},
	}
}

func resourceVmCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	network_maps := []map[string]string{}
	networks := d.Get("network").([]interface{})

	for _, network := range networks {
		net, _ := network.(map[string]interface{})

		network_maps = append(network_maps, map[string]string{
			"network_id":  net["network_id"].(string),
			"mac_address": net["mac_address"].(string),
		})
	}

	vdis := []client.VDI{}

	disks := d.Get("disk").([]interface{})

	for _, disk := range disks {
		vdi, _ := disk.(map[string]interface{})

		vdis = append(vdis, client.VDI{
			SrId:      vdi["sr_id"].(string),
			NameLabel: vdi["name_label"].(string),
			Size:      vdi["size"].(int),
		})
	}

	vm, err := c.CreateVm(
		d.Get("name_label").(string),
		d.Get("name_description").(string),
		d.Get("template").(string),
		d.Get("cloud_config").(string),
		d.Get("resource_set").(string),
		d.Get("cpus").(int),
		d.Get("memory_max").(int),
		network_maps,
		vdis,
	)

	if err != nil {
		return err
	}

	d.SetId(vm.Id)
	d.Set("cloud_config", d.Get("cloud_config").(string))
	d.Set("memory_max", d.Get("memory_max").(int))
	d.Set("resource_set", d.Get("resource_set").(string))

	vifs, err := c.GetVIFs(vm)

	if err != nil {
		return err
	}

	err = d.Set("network", vifsToMapList(vifs))

	if err != nil {
		return err
	}
	return nil
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
			"attached": disk.Attached,
			"vbd_id":   disk.Id,
			"vdi_id":   disk.VDIId,
			// "device":     disk.Device,
			// "pool_id":    disk.PoolId,
			"position":   disk.Position,
			"name_label": disk.NameLabel,
			"size":       disk.Size,
			"sr_id":      disk.SrId,
		}
		result = append(result, diskMap)
	}

	return sortDiskMapByPostion(result)
}

func vifsToMapList(vifs []client.VIF) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(vifs))
	for _, vif := range vifs {
		vifMap := map[string]interface{}{
			"attached":    vif.Attached,
			"device":      vif.Device,
			"mac_address": vif.MacAddress,
			"network_id":  vif.Network,
		}
		result = append(result, vifMap)
	}

	return sortNetworkMapByDevice(result)
}

func resourceVmRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

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
	recordToData(*vm, vifs, disks, d)
	return nil
}

func resourceVmUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

	nameLabel := d.Get("name_label").(string)
	nameDescription := d.Get("name_description").(string)
	cpus := d.Get("cpus").(int)
	autoPowerOn := d.Get("auto_poweron").(bool)
	ha := d.Get("high_availability").(string)
	rs := d.Get("resource_set").(string)
	vm, err := c.UpdateVm(d.Id(), cpus, nameLabel, nameDescription, ha, rs, autoPowerOn)
	log.Printf("[DEBUG] Retrieved vm after update: %+v\n", vm)

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

	if d.HasChange("disk") {
		oDisk, nDisk := d.GetChange("disk")

		oSet := schema.NewSet(diskHash, oDisk.([]interface{}))
		nSet := schema.NewSet(diskHash, nDisk.([]interface{}))

		removals := expandDisks(oSet.Difference(nSet).List())
		log.Printf("[DEBUG] Found the following network removals: %v previous set: %v new set: %v\n", oSet.Difference(nSet).List(), oSet, nSet)
		for _, removal := range removals {

			if err := c.DeleteDisk(*vm, removal); err != nil {
				return err
			}
		}

		additions := sortDiskByPostion(expandDisks(nSet.Difference(oSet).List()))
		log.Printf("[DEBUG] Found the following network additions: %v previous set: %v new set: %v\n", nSet.Difference(oSet).List(), oSet, nSet)
		for _, disk := range additions {

			if _, err := c.CreateDisk(*vm, disk); err != nil {
				return err
			}
		}
	}

	return resourceVmRead(d, m)
}

func resourceVmDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*client.Client)

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
				VDIId:     data["vdi_id"].(string),
				NameLabel: data["name_label"].(string),
				SrId:      data["sr_id"].(string),
				Size:      data["size"].(int),
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
		macAddress := data["mac_address"].(string)
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
	c := m.(*client.Client)

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
	recordToData(*vm, vifs, disks, d)

	return rd, nil
}

func recordToData(resource client.Vm, vifs []client.VIF, disks []client.Disk, d *schema.ResourceData) error {
	d.SetId(resource.Id)
	// d.Set("cloud_config", resource.CloudConfig)
	// err := d.Set("memory_max", resource.Memory.Size)
	// log.Printf("[DEBUG] Found error when setting memory_max %+v", err)

	// if err != nil {
	// 	return err
	// }
	d.Set("cpus", resource.CPUs.Number)
	d.Set("name_label", resource.NameLabel)
	d.Set("name_description", resource.NameDescription)
	d.Set("high_availability", resource.HA)
	d.Set("auto_poweron", resource.AutoPoweron)
	d.Set("resource_set", resource.ResourceSet)

	nets := vifsToMapList(vifs)
	err := d.Set("network", nets)

	if err != nil {
		return err
	}

	disksMapList := disksToMapList(disks)
	err = d.Set("disk", disksMapList)
	fmt.Printf("Printing disksMapList: %+v with error: %+v\n", disksMapList, err)
	if err != nil {
		return err
	}

	return nil
}

func diskHash(value interface{}) int {
	var srId string
	var nameLabel string
	var size int
	var attached bool
	switch t := value.(type) {
	case client.Disk:
		srId = t.SrId
		nameLabel = t.NameLabel
		size = t.Size
		attached = t.Attached
	case map[string]interface{}:
		srId = t["sr_id"].(string)
		nameLabel = t["name_label"].(string)
		size = t["size"].(int)
		attached = t["attached"].(bool)
	default:
		panic(fmt.Sprintf("disk cannot be hashed with type %T", t))
	}
	v := fmt.Sprintf("%s-%s-%d-%t", nameLabel, srId, size, attached)
	return hashcode.String(v)
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

	return hashcode.String(v)
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
