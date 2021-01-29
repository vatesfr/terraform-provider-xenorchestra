# xenorchestra_cloud_config

Provides information about cloud config.

## Example Usage

```hcl
data "xenorchestra_cloud_config" "cloud_config" {
  name = "Name of cloud config"
}
```

## Argument Reference
* name - (Required) The name of the cloud config you want to look up.

~> **NOTE:** If there are multiple cloud configs with the same name
Terraform will fail. Ensure that your names are unique when
using the data source.


## Attributes Reference
* id - Id of the resource set.
