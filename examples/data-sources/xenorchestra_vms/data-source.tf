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
