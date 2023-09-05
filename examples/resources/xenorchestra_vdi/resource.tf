resource "xenorchestra_vdi" "bar" {
    name_label = "alpine-virt-3-17-0"
    sr_id = data.xenorchestra_sr.sr.id
    filepath = "${path.module}/isos/alpine-virt-3.17.0-x86_64.iso"
    type = "raw"
}

# Use the vdi with the VM resource
resource "xenorchestra_vm" "bar" {

  # Other required options omitted
  # [ ... ]

  cdrom {
    id = resource.xenorchestra_vdi.bar.id
  }
}
