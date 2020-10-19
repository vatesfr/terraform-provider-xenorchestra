package xoa

import (
	"log"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var validLimitType []string = []string{"cpus", "disk", "memory"}

func resourceResourceSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceSetCreate,
		Read:   resourceSetRead,
		Update: resourceSetUpdate,
		Delete: resourceSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"subjects": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"objects": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"limit": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: StringInSlice(validLimitType, false),
						},
						"quantity": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceSetCreate(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)

	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	limits := d.Get("limit").(*schema.Set)
	objs := d.Get("objects").(*schema.Set)
	subs := d.Get("subjects").(*schema.Set)

	objects := []string{}
	for _, obj := range objs.List() {
		objects = append(objects, obj.(string))
	}
	subjects := []string{}
	for _, sub := range subs.List() {
		subjects = append(subjects, sub.(string))
	}

	rsReq := client.ResourceSet{
		Name:     name,
		Objects:  objects,
		Subjects: subjects,
	}
	for _, limit := range limits.List() {
		l := limit.(map[string]interface{})
		quantity := l["quantity"].(int)
		t := l["type"].(string)

		setLimitByType(&rsReq, t, quantity)
	}

	rs, err := c.CreateResourceSet(rsReq)

	if err != nil {
		return err
	}

	return resourceSetToData(*rs, d)
}

func resourceSetRead(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)

	if err != nil {
		return err
	}

	id := d.Id()
	rs, err := c.GetResourceSetById(id)
	log.Printf("[DEBUG] Found resource set: %+v with error: %v\n", rs, err)

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	return resourceSetToData(*rs, d)
}

func resourceSetUpdate(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)

	id := d.Id()
	rs, err := c.GetResourceSetById(id)
	if d.HasChange("limit") {
		old, new := d.GetChange("limit")

		os := old.(*schema.Set)
		ns := new.(*schema.Set)

		additions := ns.Difference(os).List()
		for _, addition := range additions {
			limit := addition.(map[string]interface{})
			t := limit["type"].(string)
			quantity := limit["quantity"].(int)
			err := c.AddResourceSetLimit(*rs, t, quantity)

			if err != nil {
				return err
			}
		}

		removals := os.Difference(ns).List()
		for _, removal := range removals {

			limit := removal.(map[string]interface{})
			t := limit["type"].(string)
			err := c.RemoveResourceSetLimit(*rs, t)

			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("subjects") {
		old, new := d.GetChange("subjects")

		os := old.(*schema.Set)
		ns := new.(*schema.Set)

		additions := ns.Difference(os).List()
		for _, addition := range additions {
			subject := addition.(string)
			err := c.AddResourceSetSubject(*rs, subject)

			if err != nil {
				return err
			}
		}

		removals := os.Difference(ns).List()
		for _, removal := range removals {
			subject := removal.(string)
			err := c.RemoveResourceSetSubject(*rs, subject)

			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("objects") {
		old, new := d.GetChange("objects")

		os := old.(*schema.Set)
		ns := new.(*schema.Set)

		additions := ns.Difference(os).List()
		for _, addition := range additions {
			subject := addition.(string)
			err := c.AddResourceSetObject(*rs, subject)

			if err != nil {
				return err
			}
		}

		removals := os.Difference(ns).List()
		for _, removal := range removals {
			object := removal.(string)
			err := c.RemoveResourceSetObject(*rs, object)

			if err != nil {
				return err
			}
		}
	}
	rs, err = c.GetResourceSetById(id)

	if err != nil {
		return err
	}

	return resourceSetToData(*rs, d)
}

func resourceSetDelete(d *schema.ResourceData, m interface{}) error {
	config := m.(client.Config)
	c, err := client.NewClient(config)

	if err != nil {
		return err
	}

	err = c.DeleteResourceSet(client.ResourceSet{Id: d.Id()})

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func resourceSetToData(rs client.ResourceSet, d *schema.ResourceData) error {
	d.SetId(rs.Id)
	d.Set("name", rs.Name)
	d.Set("subjects", rs.Subjects)
	d.Set("objects", rs.Objects)
	d.Set("limit", limitToMapList(rs.Limits))
	return nil
}

func limitToMapList(rsLimits client.ResourceSetLimits) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 3)
	if rsLimits.Cpus.Total != 0 {
		result = append(result, map[string]interface{}{
			"type":     "cpus",
			"quantity": rsLimits.Cpus.Total,
		})
	}
	if rsLimits.Disk.Total != 0 {
		result = append(result, map[string]interface{}{
			"type":     "disk",
			"quantity": rsLimits.Disk.Total,
		})
	}
	if rsLimits.Memory.Total != 0 {
		result = append(result, map[string]interface{}{
			"type":     "memory",
			"quantity": rsLimits.Memory.Total,
		})
	}

	return result
}

func setLimitByType(rs *client.ResourceSet, limitType string, limitValue int) {
	rsLimit := client.ResourceSetLimit{
		Available: limitValue,
		Total:     limitValue,
	}
	switch limitType {
	case "cpus":
		rs.Limits.Cpus = rsLimit
	case "disk":
		rs.Limits.Disk = rsLimit
	case "memory":
		rs.Limits.Memory = rsLimit
	}
}
