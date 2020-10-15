# xenorchestra_sr

Provides information about a Storage repository to ease the lookup of VM storage information.

## Example Usage

```hcl
data "xenorchestra_sr" "local_storage" {
  name_label = "Your storage repository label"
}

resource "xenorchestra_vm" "demo-vm" {
  // ...
  disk {
      sr_id = data.xenorchestra_sr.local_storage.id
      name_label = "Ubuntu Bionic Beaver 18.04_imavo"
      size = 32212254720
  }
  // ...
}
```

## Argument Reference
* name_label - (Required) The name of the storage repository you want to look up.
* pool_id - (Optional) The ID of the pool the storage repository belongs to. This is useful if you have storage repositories with the same name on different pools.
* tags - (Optional) List of tags that are applied to the storage repository.

**Note:** If there are multiple storage repositories that match terraform will fail.
Ensure that your name_label, pool_id and tags identify a unique storage repository.

## Attributes Reference
* id - Id of the storage repository.
* uuid - uuid of the storage repository. This is equivalent to the id.
* pool_id - The Id of the pool the storage repository exists on.
* sr_type - The type of storage repository (lvm, udev, iso, user, etc).
