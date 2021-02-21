# xenorchestra_vdi

Provides information about a VDI (virtual disk image)

## Example Usage

```hcl
data "xenorchestra_vdi" "vdi" {
  name_label = "ubuntu-20.04.4-live-server-amd64.iso"
}

resource "xenorchestra_vm" "demo-vm" {
  cdrom = data.xenorchestra_vdi.vdi.id
}
```

## Argument Reference
* name_label - (Required) The name of the VDI you want to look up.
* pool_id - (Optional) The ID of the pool the VDI belongs to. This is useful if you have a VDI with the same name on different pools.
* tags - (Optional) List of tags to filter down the VDI results by.

**Note:** If there are multiple VDIs that match terraform will fail.
Ensure that your name_label, pool_id and tags identify a unique VDI.

## Attributes Reference
* id - Id of the VDI.
* pool_id - The Id of the pool the VDI exists on.
