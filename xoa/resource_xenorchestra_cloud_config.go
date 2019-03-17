package xoa

import (
	"log"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCloudConfigRecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudConfigCreate,
		Read:   resourceCloudConfigRead,
		Update: resourceCloudConfigUpdate,
		Delete: resourceCloudConfigDelete,
		Importer: &schema.ResourceImporter{
			State: CloudConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCloudConfigCreate(d *schema.ResourceData, m interface{}) error {

	c, err := newXoaClient(m)

	if err != nil {
		return err
	}

	XenLog, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(XenLog)

	var resp interface{}
	params := map[string]interface{}{
		"name":     d.Get("name"),
		"template": d.Get("template"),
	}
	err = c.Call("cloudConfig.create", params, resp)

	log.Printf("cloudConfig.create %v", resp)

	if err != nil {
		return err
	}

	d.SetId("testing")
	return nil
}

func resourceCloudConfigRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCloudConfigUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceCloudConfigDelete(d *schema.ResourceData, m interface{}) error {
	c, err := newXoaClient(m)

	if err != nil {
		return err
	}

	XenLog, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(XenLog)

	var resp interface{}
	params := map[string]interface{}{
		"id": d.Id(),
	}
	err = c.Call("cloudConfig.delete", params, resp)

	log.Printf("cloudConfig.delete %v", resp)

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func CloudConfigImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}
