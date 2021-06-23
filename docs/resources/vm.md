# xenorchestra_vm

Creates a Xen Orchestra vm resource.

## Example Usage

```hcl
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
  template = <<EOF
#cloud-config

runcmd:
 - [ ls, -l, / ]
 - [ sh, -xc, "echo $(date) ': hello world!'" ]
 - [ sh, -c, echo "=========hello world'=========" ]
 - ls -l /root
EOF
}

resource "xenorchestra_vm" "bar" {
    memory_max = 1073733632
    cpus  = 1
    cloud_config = xenorchestra_cloud_config.bar.template
    name_label = "Name"
    name_description = "description"
    template = data.xenorchestra_template.template.id

    # Prefer to run the VM on the primary pool instance
    affinity_host = data.xenorchestra_pool.pool.master
    network {
      network_id = data.xenorchestra_network.net.id
    }

    disk {
      sr_id = "7f469400-4a2b-5624-cf62-61e522e50ea1"
      name_label = "Ubuntu Bionic Beaver 18.04_imavo"
      size = 32212254720 
    }

    tags = [
      "Ubuntu",
      "Bionic",
    ]
}
```

## Argument Reference
* `name_label` - (Required) The name of VM.
* `name_description` - (Optional) The description of the VM.
* `template` - (Required) The ID of the VM template to create the new VM from.
* `cloud_config` - (Optional) The content of the cloud-init config to use
* `cloud_network_config` - (Optional) The content of the cloud-init network configuration for the VM (uses [version 1](https://cloudinit.readthedocs.io/en/latest/topics/network-config-format-v1.html))
* `cpus` - (Required) The number of CPUs the VM will have. Updates to this field will cause a stop and start of the VM if the new CPU value is greater than the max CPU value. This can be determined with the following command:
```
$ xo-cli xo.getAllObjects filter='json:{"id": "cf7b5d7d-3cd5-6b7c-5025-5c935c8cd0b8"}' | jq '.[].CPUs'
{
  "max": 4,
  "number": 2
}

# Updating the VM to use 3 CPUs would happen without stopping/starting the VM
# Updating the VM to use 5 CPUs would stop/start the VM
```
* `memory_max` - (Required) The amount of memory in bytes the VM will have. Updates to this field will cause a stop and start of the VM if the new `memory_max` value is greater than the dynamic memory max. This can be determined with the following command:
```
$ xo-cli xo.getAllObjects filter='json:{"id": "cf7b5d7d-3cd5-6b7c-5025-5c935c8cd0b8"}' | jq '.[].memory.dynamic'
[
  2147483648, # memory dynamic min
  4294967296  # memory dynamic max (4GB)
]
# Updating the VM to use 3GB of memory would happen without stopping/starting the VM
# Updating the VM to use 5GB of memory would stop/start the VM
```


* `high_availabililty` - (Optional) The restart priority for the VM. Possible values are `best-effort`, `restart` and empty string (no restarts on failure. Defaults to empty string.
* `installation_method` - (Optional) This cannot be used with `cdrom`. Possible values are `network` which allows a VM to boot via PXE.
* `auto_poweron` - (Optional) If the VM will automatically turn on. Defaults to `false`.
* `affinity_host` - (Optional) The preferred host you would like the VM to run on. If changed on an existing VM it will require a reboot for the VM to be rescheduled.
* `wait_for_ip` - (Optional) Whether terraform should wait until IP addresses are present on the VM's network interfaces before considering it created. This only works if guest-tools are installed in the VM. Defaults to false.
* `cdrom` - (Optional) The ISO that should be attached to VM. This allows you to create a VM from a diskless template (any templates available from `xe template-list`) and install the OS from the following ISO.
    * `id` - The ID of the ISO (VDI) to attach to the VM. This can be easily provided by using the `vdi` data source.
* `network` - (Required) The network the VM will use
    * `network_id` - (Required) The ID of the network the VM will be on.
    * `mac_address` - (Optional) The mac address of the network interface. This must be parsable by go's [net.ParseMAC function](https://golang.org/pkg/net/#ParseMAC). All mac addresses are stored in Terraform's state with [HardwareAddr's string representation](https://golang.org/pkg/net/#HardwareAddr.String) i.e. 00:00:5e:00:53:01
* `disk` - (Required) The disk the VM will have access to.
    * `sr_id` - (Required) The storage repository ID to use.
    * `name_label` - (Required) The name for the disk.
    * `name_description` - (Optional) A description for the disk.
    * `size` - (Required) The size in bytes of the disk.
* `tags` - (Optional) List of labels (strings) that are used to identify and organize resources. These are equivalent to Xenserver [tags](https://docs.citrix.com/en-us/xencenter/7-1/resources-tagging.html).

## Attributes Reference
In addition to all the arguments above, the following attributes are exported:

* `id` - the ID generated by the XO api.
* `power_state` - the power state of the VM. This can be Running, Halted, Paused or Suspended.
* `ipv4_addresses` - This is only accessible if guest-tools is installed in the VM and if `wait_for_ip` is set to true. This will contain a list of the ipv4 addresses across all network interfaces in order
* `ipv6_addresses` - This is only accessible if guest-tools is installed in the VM and if `wait_for_ip` is set to true. This will contain a list of the ipv6 addresses across all network interfaces in order.

For `network` blocks the following attributes are exported:
* `ipv4_addresses` - This is only accessible if guest-tools is installed in the VM and if `wait_for_ip` is set to true. This will contain a list of the ipv4 addresses for the specific network interface. See the example below for more details.
* `ipv6_addresses` - This is only accessible if guest-tools is installed in the VM and if `wait_for_ip` is set to true. This will contain a list of the ipv6 addresses for the specific network interface. See the example below for more details.

```hcl
resource "xenorchestra_vm" "vm" {
  ...

  # Specify VM with two network interfaces
  network {
    ...
  }
  network {
    ...
  }
}

output "first-network-interface-ips" {
  value = xenorchestra_vm.vm.network[0].ipv4_addresses
}

output "all-ipv4-ips-for-vm" {
  value = xenorchestra_vm.vm.ipv4_addresses
  # This is also the same as xenorchestra_vm.vm.network[*].ipv4_addresses
}
```
