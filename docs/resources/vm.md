---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "xenorchestra_vm Resource - terraform-provider-xenorchestra"
subcategory: ""
description: |-
  Creates a Xen Orchestra vm resource.
  Differences with the Xen Orchestra UI
  Cloudinit
  Xen Orchestra allows templating cloudinit config through its own custom mechanism:
  * "{name}" is replaced with the VM's name
  * "%" is replaced with the VM's index
  This does not work in terraform since that is applied on Xen Orchestra's client side (Javascript). Terraform provides a "templatefile" function that allows for a similar substitution. Please see the example below for more details.
---

# xenorchestra_vm (Resource)

Creates a Xen Orchestra vm resource.

## Differences with the Xen Orchestra UI

### Cloudinit

Xen Orchestra allows templating cloudinit config through its own custom mechanism:
* "{name}" is replaced with the VM's name
* "%" is replaced with the VM's index

This does not work in terraform since that is applied on Xen Orchestra's client side (Javascript). Terraform provides a "templatefile" function that allows for a similar substitution. Please see the example below for more details.

## Example Usage

```terraform
/*
# cloud_config.tftpl file used by the cloudinit templating.

#cloud-config
hostname: ${hostname}
fqdn: ${hostname}.${domain}
package_upgrade: true
*/

# Content of the terraform files
data "xenorchestra_pool" "pool" {
  name_label = "pool name"
}

data "xenorchestra_template" "template" {
  name_label = "Puppet agent - Bionic 18.04 - 1"
}

data "xenorchestra_network" "net" {
  name_label = "Pool-wide network associated with eth0"
}

resource "xenorchestra_cloud_config" "bar" {
  name = "cloud config name"
  # Template the cloudinit if needed
  template = templatefile("cloud_config.tftpl", {
    hostname = "your-hostname"
    domain   = "your.domain.com"
  })
}

resource "xenorchestra_vm" "bar" {
  memory_max       = 1073733632
  cpus             = 1
  cloud_config     = xenorchestra_cloud_config.bar.template
  name_label       = "Name"
  name_description = "description"
  template         = data.xenorchestra_template.template.id

  # Prefer to run the VM on the primary pool instance
  affinity_host = data.xenorchestra_pool.pool.master
  network {
    network_id = data.xenorchestra_network.net.id
  }

  disk {
    sr_id      = "7f469400-4a2b-5624-cf62-61e522e50ea1"
    name_label = "Ubuntu Bionic Beaver 18.04_imavo"
    size       = 32212254720
  }

  tags = [
    "Ubuntu",
    "Bionic",
  ]

  // Override the default create timeout from 5 mins to 20.
  timeouts {
    create = "20m"
  }

  // Note: Xen Orchestra populates values within Xenstore and will need ignored via
  // lifecycle ignore_changes or modeled in your terraform code
  xenstore = {
    key1 = "val1"
    key2 = "val2"
  }
}

# vm resource that waits until its first network interface
# is assigned an IP via DHCP
resource "xenorchestra_vm" "vm" {
  # Specify VM with two network interfaces
  network {
    network_id       = "7ed8998b-405c-40b5-b164-f9058efcf6b4"
    expected_ip_cidr = "10.0.0.0/16"
  }
  network {
    network_id = "b4cf8532-ae43-493b-9fc6-a6b456d16876"
  }
  # required arguments for xenorchestra_vm
  cpus = 2
  disk {
    sr_id      = "7f469400-4a2b-5624-cf62-61e522e50ea1"
    name_label = "Ubuntu Bionic Beaver 18.04_imavo"
    size       = 32212254720
  }
  memory_max = 1073733632
  name_label = "Name"
  template   = data.xenorchestra_template.template.id
}

output "first-network-interface-ips" {
  value = xenorchestra_vm.vm.network[0].ipv4_addresses
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cpus` (Number) The number of CPUs the VM will have. Updates to this field will cause a stop and start of the VM if the new CPU value is greater than the max CPU value. This can be determined with the following command:
```

$ xo-cli xo.getAllObjects filter='json:{"id": "cf7b5d7d-3cd5-6b7c-5025-5c935c8cd0b8"}' | jq '.[].CPUs'
{
  "max": 4,
  "number": 2
}

# Updating the VM to use 3 CPUs would happen without stopping/starting the VM
# Updating the VM to use 5 CPUs would stop/start the VM
```
- `disk` (Block List, Min: 1) The disk the VM will have access to. (see [below for nested schema](#nestedblock--disk))
- `memory_max` (Number) The amount of memory in bytes the VM will have. Updates to this field will case a stop and start of the VM if the new value is greater than the dynamic memory max. This can be determined with the following command:
```


$ xo-cli xo.getAllObjects filter='json:{"id": "cf7b5d7d-3cd5-6b7c-5025-5c935c8cd0b8"}' | jq '.[].memory.dynamic'
[
  2147483648, # memory dynamic min
  4294967296  # memory dynamic max (4GB)
]
# Updating the VM to use 3GB of memory would happen without stopping/starting the VM
# Updating the VM to use 5GB of memory would stop/start the VM
```
- `name_label` (String) The name of the VM.
- `network` (Block List, Min: 1) The network for the VM. (see [below for nested schema](#nestedblock--network))
- `template` (String) The ID of the VM template to create the new VM from.

### Optional

- `affinity_host` (String) The preferred host you would like the VM to run on. If changed on an existing VM it will require a reboot for the VM to be rescheduled.
- `auto_poweron` (Boolean) If the VM will automatically turn on. Defaults to `false`.
- `blocked_operations` (Set of String) List of operations on a VM that are not permitted. Examples include: clean_reboot, clean_shutdown, hard_reboot, hard_shutdown, pause, shutdown, suspend, destroy. See: https://xapi-project.github.io/xen-api/classes/vm.html#enum_vm_operations
- `cdrom` (Block List, Max: 1) The ISO that should be attached to VM. This allows you to create a VM from a diskless template (any templates available from `xe template-list`) and install the OS from the following ISO. (see [below for nested schema](#nestedblock--cdrom))
- `clone_type` (String) The type of clone to perform for the VM. Possible values include `fast` or `full` and defaults to `fast`. In order to perform a `full` clone, the VM template must not be a disk template.
- `cloud_config` (String) The content of the cloud-init config to use. See the cloud init docs for more [information](https://cloudinit.readthedocs.io/en/latest/topics/examples.html).
- `cloud_network_config` (String) The content of the cloud-init network configuration for the VM (uses [version 1](https://cloudinit.readthedocs.io/en/latest/topics/network-config-format-v1.html))
- `core_os` (Boolean)
- `cpu_cap` (Number)
- `cpu_weight` (Number)
- `destroy_cloud_config_vdi_after_boot` (Boolean) Determines whether the cloud config VDI should be deleted once the VM has booted. Defaults to `false`. If set to `true`, power_state must be set to `Running`.
- `exp_nested_hvm` (Boolean) Boolean parameter that allows a VM to use nested virtualization.
- `high_availability` (String) The restart priority for the VM. Possible values are `best-effort`, `restart` and empty string (no restarts on failure. Defaults to empty string
- `host` (String)
- `hvm_boot_firmware` (String) The firmware to use for the VM. Possible values are `bios` and `uefi`.
- `installation_method` (String) This cannot be used with `cdrom`. Possible values are `network` which allows a VM to boot via PXE.
- `name_description` (String) The description of the VM.
- `power_state` (String) The power state of the VM. This can be Running, Halted, Paused or Suspended.
- `resource_set` (String)
- `start_delay` (Number) Number of seconds the VM should be delayed from starting.
- `tags` (Set of String) The tags (labels) applied to the given entity.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `vga` (String) The video adapter the VM should use. Possible values include std and cirrus.
- `videoram` (Number) The videoram option the VM should use. Possible values include 1, 2, 4, 8, 16
- `xenstore` (Map of String) The key value pairs to be populated in xenstore.

### Read-Only

- `id` (String) The ID of this resource.
- `ipv4_addresses` (List of String) This is only accessible if guest-tools is installed in the VM. While the output contains a list of ipv4 addresses, the presence of an IP address is only guaranteed if `expected_ip_cidr` is set for that interface. The list contains the ipv4 addresses across all network interfaces in order. See the example terraform code for more details.
- `ipv6_addresses` (List of String) This is only accessible if guest-tools is installed in the VM. While the output contains a list of ipv6 addresses, the presence of an IP address is only guaranteed if `expected_ip_cidr` is set for that interface. The list contains the ipv6 addresses across all network interfaces in order.

<a id="nestedblock--disk"></a>
### Nested Schema for `disk`

Required:

- `name_label` (String) The name for the disk
- `size` (Number) The size in bytes for the disk.
- `sr_id` (String) The storage repository ID to use.

Optional:

- `attached` (Boolean) Whether the device should be attached to the VM.
- `name_description` (String) The description for the disk

Read-Only:

- `position` (String) Indicates the order of the block device.
- `vbd_id` (String)
- `vdi_id` (String)


<a id="nestedblock--network"></a>
### Nested Schema for `network`

Required:

- `network_id` (String) The ID of the network the VM will be on.

Optional:

- `attached` (Boolean) Whether the device should be attached to the VM.
- `expected_ip_cidr` (String) Determines the IP CIDR range the provider will wait for on this network interface. Resource creation is not complete until an IP address within the specified range becomes available. This parameter replaces the former `wait_for_ip` functionality. This only works if guest-tools are installed in the VM. Defaults to "", which skips IP address matching.
- `mac_address` (String) The mac address of the network interface. This must be parsable by go's [net.ParseMAC function](https://golang.org/pkg/net/#ParseMAC). All mac addresses are stored in Terraform's state with [HardwareAddr's string representation](https://golang.org/pkg/net/#HardwareAddr.String) i.e. 00:00:5e:00:53:01

Read-Only:

- `device` (String)
- `ipv4_addresses` (List of String)
- `ipv6_addresses` (List of String)


<a id="nestedblock--cdrom"></a>
### Nested Schema for `cdrom`

Required:

- `id` (String) The ID of the ISO (VDI) to attach to the VM. This can be easily provided by using the `vdi` data source.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)
