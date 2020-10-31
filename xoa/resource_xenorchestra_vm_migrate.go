package xoa

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func vmResourceV0() *schema.Resource {
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
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attached": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
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
				Set: func(value interface{}) int {
					network := value.(map[string]interface{})

					macAddress := network["mac_address"].(string)
					networkId := network["network_id"].(string)
					v := fmt.Sprintf("%s-%s", macAddress, networkId)
					log.Printf("[TRACE] Setting network via %s\n", v)

					return hashcode.String(v)
				},
			},
			"disk": &schema.Schema{
				Type:     schema.TypeSet,
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
						"size": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
				Set: func(value interface{}) int {
					var buf bytes.Buffer
					disk := value.(map[string]interface{})

					buf.WriteString(fmt.Sprintf("%s-", disk["sr_id"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", disk["name_label"].(string)))
					buf.WriteString(fmt.Sprintf("%d-", disk["size"]))
					return hashcode.String(buf.String())
				},
			},
		},
	}
}

func migrateXOVmStateV0ToV1(rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {

	log.Println("[INFO] Attempting to migrate VM terraform state to V1")

	networks, ok := rawState["network"].([]interface{})

	if !ok {
		return rawState, errors.New(fmt.Sprintf("failed to coerce 'network' on VM state migration to V1. Found %+v with %T", rawState["network"], rawState["network"]))
	}

	networkMap := []map[string]interface{}{}
	for _, network := range networks {

		networkMap = append(networkMap, network.(map[string]interface{}))
	}
	newNetworks := sortNetworksByDevice(networkMap)
	rawState["network"] = newNetworks
	return rawState, nil
}
