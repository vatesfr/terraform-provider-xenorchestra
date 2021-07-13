# xenorchestra_vms

Use this data source to filter Xenorchestra VMs by certain criteria (pool_id, power_state or host) for use in other resources.

## Example Usage

```hcl
data "xenorchestra_pool" "pool" {
  name_label = "Your pool"
}

data "xenorchestra_vms" "vms" {
  pool_id = data.xenorchestra_pool.pool.id
  power_state = "Running"
  host = data.xenorchestra_pool.pool.master
}

output "vms_max_memory_map" {
  value = tomap({
  for k, vm in data.xenorchestra_vms.vms.vms : k => vm.memory_max
  })
}
output "vms_length" {
  value = length(data.xenorchestra_vms.vms.vms)
}
```

## Argument Reference

* pool_id - (Required) The ID of the pool the vms belong to.
* host - (Optional) The ID of the host (container) the vms belong to.
* power_state - (Optional) The power state of the vms ("Running" / "Halted")

## Attributes Reference

* id - The CRC-32 checksum based on arguments passed to this data source.
* pool_id - The ID of the pool the VM belongs to.
* vms - A list of information for all vms found in this pool.
    * vms.id - The uuid of the VM.
    * vms.name_label - The name label of the VM.
    * vms.cpu - The number of cpu's of the VM.
    * vms.cloud_config - The cloud configuration for this VM.
    * vms.cloud_network_config - The cloud network configuration for this VM.
    * vms.tags - The tags applied to the VM.
    * vms.memory_max - The maximum memory size of the VM.
    * vms.affinity_host - The affinity host of the VM.
    * vms.template - The template used to create the VM.
    * vms.wait_for_ip - The wait for ip option of the VM.
    * vms.high_availability - The high availability option of the VM.
    * vms.resource_set - The resource set of this VM.
    * vms.ipv4_addresses - A list of ipv4 addresses of the vm.
    * vms.ipv6_addresses - A list of ipv6 addresses of the VM.