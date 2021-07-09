# xenorchestra_host

Provides information about a host.

## Example Usage

```hcl
data "xenorchestra_host" "host1" {
  name_label = "Your host"
}
resource "xenorchestra_vm" "node" {
    //...
    affinity_host = data.xenorchestra_host.ng6.id
    //...
}

```

## Argument Reference
* name_label - (Required) The name of the host you want to look up.

~> **NOTE:** If there are multiple hosts with the same name
Terraform will fail. Ensure that your names are unique when
using the data source.

## Attributes Reference
* id - The id of the host.
* name_label - Name label of the host.
* pool_id - Id of the pool that the host belongs to.
* tags - The tags applied to the host.
* memory - The memory size for the host.
* memory_usage - The memory usage for the host.
* cpus - Host cpu information.
    * cores - The number of cores.
    * sockets - The number of sockets.
