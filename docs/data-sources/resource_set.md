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

## Attributes Reference
* id - Id of the resource set.
