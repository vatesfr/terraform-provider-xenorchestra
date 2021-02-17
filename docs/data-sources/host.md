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
* id - Id of the host.
