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

## Attributes Reference
* id - Id of the host.
