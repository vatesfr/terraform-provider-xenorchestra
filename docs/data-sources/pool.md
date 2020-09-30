# xenorchestra_pool

Provides information about a pool.

## Example Usage

```hcl
data "xenorchestra_pool" "pool" {
  name_label = "Your pool"
}
data "xenorchestra_sr" "local_storage" {
  name_label = "Your storage repository label"
  pool_id = data.xenorchestra_pool.pool.id
}
```

## Argument Reference
* name_label - (Required) The name of the pool you want to look up.

## Attributes Reference
* id - Id of the pool.
* description - The description of the pool.
* cpus - CPU information about the pool.
    * cores - Number of cores in the pool.
    * sockets - Number of sockets in the pool.
