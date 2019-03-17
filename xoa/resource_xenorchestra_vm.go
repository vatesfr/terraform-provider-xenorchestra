package xoa

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/powerman/rpc-codec/jsonrpc2"
)

var logFile = "/tmp/terraform-provider-xenorchestra"
var XenLog io.Writer
var err error

func init() {
}

func resourceRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceVmCreate,
		Read:   resourceVmRead,
		Update: resourceVmUpdate,
		Delete: resourceVmDelete,
		Importer: &schema.ResourceImporter{
			State: RecordImport,
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
			"template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cloud_config": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
			// "existingDisks": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// },
			// "VIFs": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// },
		},
	}
}

func resourceVmCreate(d *schema.ResourceData, m interface{}) error {
	return nil
}

type xoaConfig struct {
	host     string
	username string
	password string
}

func resourceVmRead(d *schema.ResourceData, m interface{}) error {
	xoaId := d.Id()
	c, err := newXoaClient(m)

	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"limit": 1000,
		"type":  "VM",
	}
	var objsRes allObjectResponse
	err = c.Call("xo.getAllObjects", params, &objsRes.Objects)
	if err != nil {
		return err
	}
	vmObj := objsRes.Objects[xoaId]
	log.Printf("vmobj: %v", vmObj)
	recordToData(vmObj, d)
	return nil
}

func resourceVmUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceVmDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func RecordImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	xoaId := d.Id()

	var c *rpc.Client
	c, err = newXoaClient(m)

	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"limit": 1000,
		"type":  "VM",
	}
	var objsRes allObjectResponse
	err = c.Call("xo.getAllObjects", params, &objsRes.Objects)
	if err != nil {
		return nil, err
	}
	vmObj := objsRes.Objects[xoaId]
	recordToData(vmObj, d)
	return []*schema.ResourceData{d}, nil
}

func newXoaClient(m interface{}) (*rpc.Client, error) {
	creds := m.(xoaConfig)
	ws, _, err := dialer.Dial(fmt.Sprintf("ws://%s/api/", creds.host), http.Header{})

	if err != nil {
		return nil, err
	}

	codec := jsonrpc2.NewClientCodec(&rwc{c: ws})
	c := rpc.NewClientWithCodec(codec)

	params := map[string]interface{}{
		"email":    creds.username,
		"password": creds.password,
	}
	var reply clientResponse
	err = c.Call("session.signInWithPassword", params, &reply)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func recordToData(resource VmObject, d *schema.ResourceData) error {
	d.SetId(resource.Id)
	d.Set("cloud_config", resource.CloudConfig)
	d.Set("memory_max", resource.Memory.Size)
	d.Set("cpus", resource.CPUs.Number)
	d.Set("name_label", resource.NameLabel)
	d.Set("name_description", resource.NameDescription)
	return nil
}

type CPUs struct {
	Number int
}

type MemoryObject struct {
	Dynamic []int `json:"dynamic"`
	Static  []int `json:"static"`
	Size    int   `json:"size"`
}

type VmObject struct {
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

type allObjectResponse struct {
	Objects map[string]VmObject `json:"-"`
}
