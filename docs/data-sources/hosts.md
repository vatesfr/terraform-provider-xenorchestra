# xenorchestra_hosts

Use this data source to filter Xenorchestra hosts by certain criteria (name_label, tags) for use in other resources.

## Example Usage

```hcl
data "xenorchestra_hosts" "hosts" {
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
  count = length(data.xenorchestra_hosts.hosts)

  affinity_host = data.xenorchestra_hosts.hosts[count.index].id
  ...
  ...
}
```

## Argument Reference
* pool_id - (Required) The pool id used to filter the resulting hosts by
* tags - (Optional) The tags used to filter the resulting hosts by
* sorted_by - (Optional) The host field you would like to sort by (id and name_label supported)
* sort_order - (Optional) Valid options are "asc" or "desc" and the sort order is applied to the `sorted_by` argument

## Attributes Reference
* master - The primary host of the pool
* hosts - List containing the matching hosts after applying the argument filtering. 
  * id - The id of the host.
  * name_label - Name label of the host.
  * pool_id - Id of the pool that the host belongs to.
  * tags - The tags applied to the host.
  * cpus - Host cpu information.
    * cores - The number of cores. 
    * sockets - The number of sockets.