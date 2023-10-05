package state

import (
	"context"
	"fmt"
	"log"
	"net"

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

var defaultCloudConfigDiskName string = "XO CloudConfigDrive"

func VmStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	rawState["destroy_cloud_config_vdi_after_boot"] = false
	return rawState, nil
}

func ResourceVmResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{

			"affinity_host": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The preferred host you would like the VM to run on. If changed on an existing VM it will require a reboot for the VM to be rescheduled.",
				Optional:    true,
			},
			"blocked_operations": &schema.Schema{
				Type:        schema.TypeSet,
				Description: "List of operations on a VM that are not permitted. Examples include: clean_reboot, clean_shutdown, hard_reboot, hard_shutdown, pause, shutdown, suspend, destroy. This can be used to prevent a VM from being destroyed. The entire list can be found here",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name_label": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the VM.",
				Required:    true,
			},
			"name_description": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The description of the VM.",
				Optional:    true,
			},
			"cloud_network_config": &schema.Schema{
				Description: "The content of the cloud-init network configuration for the VM (uses [version 1](https://cloudinit.readthedocs.io/en/latest/topics/network-config-format-v1.html))",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"auto_poweron": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "If the VM will automatically turn on. Defaults to `false`.",
				Default:     false,
				Optional:    true,
			},
			"exp_nested_hvm": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Boolean parameter that allows a VM to use nested virtualization.",
				Default:     false,
				Optional:    true,
			},
			"hvm_boot_firmware": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "bios",
				Optional:     true,
				Description:  "The firmware to use for the VM. Possible values are `bios` and `uefi`.",
				ValidateFunc: validation.StringInSlice(validFirmware, false),
			},
			"power_state": &schema.Schema{
				Description: "The power state of the VM. This can be Running, Halted, Paused or Suspended.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"installation_method": &schema.Schema{
				Type:          schema.TypeString,
				Description:   "This cannot be used with `cdrom`. Possible values are `network` which allows a VM to boot via PXE.",
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(validInstallationMethods, false),
				ConflictsWith: []string{"cdrom"},
			},
			"high_availability": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "",
				Description:  "The restart priority for the VM. Possible values are `best-effort`, `restart` and empty string (no restarts on failure. Defaults to empty string",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validHaOptions, false),
			},
			"template": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The ID of the VM template to create the new VM from.",
				Required:    true,
				ForceNew:    true,
			},
			"cloud_config": &schema.Schema{
				Description: "The content of the cloud-init config to use. See the cloud init docs for more [information](https://cloudinit.readthedocs.io/en/latest/topics/examples.html).",
				Type:        schema.TypeString,
				Optional:    true,
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
				Description: `The number of CPUs the VM will have. Updates to this field will cause a stop and start of the VM if the new CPU value is greater than the max CPU value. This can be determined with the following command:

$ xo-cli xo.getAllObjects filter='json:{"id": "cf7b5d7d-3cd5-6b7c-5025-5c935c8cd0b8"}' | jq '.[].CPUs'
{
  "max": 4,
  "number": 2
}
		    
# Updating the VM to use 3 CPUs would happen without stopping/starting the VM
# Updating the VM to use 5 CPUs would stop/start the VM`,
			},
			"memory_max": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				Description: `The amount of memory in bytes the VM will have. Updates to this field will case a stop and start of the VM if the new value is greater than the dynamic memory max. This can be determined with the following command:

$ xo-cli xo.getAllObjects filter='json:{"id": "cf7b5d7d-3cd5-6b7c-5025-5c935c8cd0b8"}' | jq '.[].memory.dynamic'
[
  2147483648, # memory dynamic min
  4294967296  # memory dynamic max (4GB)
]
# Updating the VM to use 3GB of memory would happen without stopping/starting the VM
# Updating the VM to use 5GB of memory would stop/start the VM
			`,
			},
			"resource_set": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipv4_addresses": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "This is only accessible if guest-tools is installed in the VM and if `wait_for_ip` is set to true. This will contain a list of the ipv4 addresses across all network interfaces in order.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ipv6_addresses": &schema.Schema{
				Type:        schema.TypeList,
				Description: "This is only accessible if guest-tools is installed in the VM and if `wait_for_ip` is set to true. This will contain a list of the ipv6 addresses across all network interfaces in order.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vga": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "std",
				Description:  "The video adapter the VM should use. Possible values include std and cirrus.",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(validVga, false),
			},
			"videoram": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "The videoram option the VM should use. Possible values include 1, 2, 4, 8, 16",
				Default:     8,
				Optional:    true,
			},
			"start_delay": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "Number of seconds the VM should be delayed from starting.",
				Default:     0,
				Optional:    true,
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
				Type:        schema.TypeBool,
				Default:     false,
				Description: "Whether terraform should wait until IP addresses are present on the VM's network interfaces before considering it created. This only works if guest-tools are installed in the VM. Defaults to false.",
				Optional:    true,
			},
			"cdrom": &schema.Schema{
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "The ISO that should be attached to VM. This allows you to create a VM from a diskless template (any templates available from `xe template-list`) and install the OS from the following ISO.",
				ConflictsWith: []string{"installation_method"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Description: "The ID of the ISO (VDI) to attach to the VM. This can be easily provided by using the `vdi` data source.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
				// This should be increased but I don't understand
				// the use cases for multiple ISOs just yet. For now
				// limit it to a single ISO
				MaxItems: 1,
			},
			"network": &schema.Schema{
				Type:        schema.TypeList,
				Description: "The network for the VM.",
				Required:    true,
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
							Description: "The ID of the network the VM will be on.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"mac_address": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The mac address of the network interface. This must be parsable by go's [net.ParseMAC function](https://golang.org/pkg/net/#ParseMAC). All mac addresses are stored in Terraform's state with [HardwareAddr's string representation](https://golang.org/pkg/net/#HardwareAddr.String) i.e. 00:00:5e:00:53:01",
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
				Type:        schema.TypeList,
				Required:    true,
				Description: "The disk the VM will have access to.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sr_id": &schema.Schema{
							Description: "The storage repository ID to use.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"name_label": &schema.Schema{
							Description: "The name for the disk",
							Type:        schema.TypeString,
							Required:    true,
						},
						"name_description": &schema.Schema{
							Description: "The description for the disk",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"size": &schema.Schema{
							Description: "The size in bytes for the disk.",
							Type:        schema.TypeInt,
							Required:    true,
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
			"tags": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The tags (labels) applied to the given entity.",
			},
		},
	}
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
