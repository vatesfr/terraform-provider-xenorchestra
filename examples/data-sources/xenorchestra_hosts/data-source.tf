data "xenorchestra_hosts" "pool1" {
  pool_id = data.xenorchestra_pool.pool.id

  sort_by = "name_label"
  sort_order = "asc"

  # Optionally filter by tags if needed
  tags = [
    "tag1",
    "tag2",
  ]
}

resource "xenorchestra_vm" "vm" {
  count = length(data.xenorchestra_hosts.pool1.hosts)

  affinity_host = data.xenorchestra_hosts.pool1.hosts[count.index].id
  ...
  ...
}
