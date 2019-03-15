package xoa

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"os"

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
			// "template": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// 	ForceNew: true,
			// },
			"cloudConfig": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			// "coreOs": &schema.Schema{
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Default:  false,
			// },
			// "cpuCap": &schema.Schema{
			// 	Type:     schema.TypeInt,
			// 	Optional: true,
			// 	Default:  0,
			// },
			// "cpuWeight": &schema.Schema{
			// 	Type:     schema.TypeInt,
			// 	Optional: true,
			// 	Default:  0,
			// },
			"CPUs": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"memoryMax": &schema.Schema{
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

	XenLog, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(XenLog)

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
	// d.Set("cloudConfig", resource.cloudConfig)
	d.Set("memoryMax", resource.Memory.Size)
	d.Set("CPUs", resource.CPUs.Number)
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
	// Template    string `json:"template"`
	// CloudConfig string `json:"cloudConfig"`
}

type allObjectResponse struct {
	Objects map[string]VmObject `json:"-"`
}
