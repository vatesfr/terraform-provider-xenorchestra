package xoa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

var validLimitType []string = []string{"cpus", "disk", "memory"}

func resourceResourceSet() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates a Xen Orchestra resource set.",
		CreateContext: resourceSetCreateContext,
		ReadContext:   resourceSetReadContext,
		UpdateContext: resourceSetUpdateContext,
		DeleteContext: resourceSetDeleteContext,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the resource set.",
			},
			"subjects": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The uuids of the user accounts that should have access to the resource set.",
			},
			"objects": &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The uuids of the objects that are within scope of the resource set. A minimum of a storage repository, network and VM template are required for users to launch VMs.",
			},
			"limit": &schema.Schema{
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The limit applied to the resource set.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:         schema.TypeString,
							Description:  "The type of resource set limit. Must be cpus, memory or disk.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(validLimitType, false),
						},
						"quantity": &schema.Schema{
							Type:        schema.TypeInt,
							Description: "The numerical limit for the given type.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceSetCreateContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

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
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceSetToData(*rs, d))
}

func resourceSetReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	id := d.Id()
	rs, err := c.GetResourceSetById(id)
	tflog.Debug(ctx, "Found resource set", map[string]interface{}{
		"resource_set": rs,
		"error":        err,
	})

	if _, ok := err.(client.NotFound); ok {
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceSetToData(*rs, d))
}

func resourceSetUpdateContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	id := d.Id()
	rs, _ := c.GetResourceSetById(id)
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
				return diag.FromErr(err)
			}
		}

		removals := os.Difference(ns).List()
		for _, removal := range removals {

			limit := removal.(map[string]interface{})
			t := limit["type"].(string)
			err := c.RemoveResourceSetLimit(*rs, t)

			if err != nil {
				return diag.FromErr(err)
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
				return diag.FromErr(err)
			}
		}

		removals := os.Difference(ns).List()
		for _, removal := range removals {
			subject := removal.(string)
			err := c.RemoveResourceSetSubject(*rs, subject)

			if err != nil {
				return diag.FromErr(err)
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
				return diag.FromErr(err)
			}
		}

		removals := os.Difference(ns).List()
		for _, removal := range removals {
			object := removal.(string)
			err := c.RemoveResourceSetObject(*rs, object)

			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	rs, err := c.GetResourceSetById(id)

	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(resourceSetToData(*rs, d))
}

func resourceSetDeleteContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(client.XOClient)

	err := c.DeleteResourceSet(client.ResourceSet{Id: d.Id()})

	if err != nil {
		return diag.FromErr(err)
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
