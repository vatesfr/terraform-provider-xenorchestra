# xenorchestra_resource_set

Provides information about a resource set.

## Example Usage

```hcl
data "xenorchestra_resource_set" "rs" {
  name = "my resource set"
}
```

## Argument Reference
* name - (Required) The name of the resource set you want to look up.

~> **NOTE:** If there are multiple resource sets with the same name
Terraform will fail. Ensure that your resource set names are unique when
using the data source.


## Attributes Reference
* id - Id of the resource set.
