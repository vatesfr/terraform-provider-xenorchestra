package xoa

import (
	"io"
	"log"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform/helper/schema"
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
	c, err := client.NewClient()

	if err != nil {
		return err
	}

	vm, err := c.CreateVm(
		d.Get("name_label").(string),
		d.Get("name_description").(string),
		d.Get("template").(string),
		d.Get("cloud_config").(string),
		d.Get("cpus").(int),
		d.Get("memory_max").(int),
	)

	if err != nil {
		return err
	}

	recordToData(*vm, d)
	return nil
}

type xoaConfig struct {
	host     string
	username string
	password string
}

func resourceVmRead(d *schema.ResourceData, m interface{}) error {
	xoaId := d.Id()
	c, err := client.NewClient()

	if err != nil {
		return err
	}
	vmObj, err := c.GetVm(xoaId)
	if err != nil {
		return err
	}
	log.Printf("vmobj: %v", vmObj)
	recordToData(*vmObj, d)
	return nil
}

func resourceVmUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceVmDelete(d *schema.ResourceData, m interface{}) error {
	c, err := client.NewClient()

	if err != nil {
		return err
	}

	err = c.DeleteVm(d.Id())

	if err != nil {
		return err
	}
	return nil
}

func RecordImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	xoaId := d.Id()

	c, err := client.NewClient()

	if err != nil {
		return nil, err
	}

	vmObj, err := c.GetVm(xoaId)
	if err != nil {
		return nil, err
	}
	recordToData(*vmObj, d)
	return []*schema.ResourceData{d}, nil
}

func recordToData(resource client.Vm, d *schema.ResourceData) error {
	d.SetId(resource.Id)
	d.Set("cloud_config", resource.CloudConfig)
	d.Set("memory_max", resource.Memory.Size)
	d.Set("cpus", resource.CPUs.Number)
	d.Set("name_label", resource.NameLabel)
	d.Set("name_description", resource.NameDescription)
	return nil
}
