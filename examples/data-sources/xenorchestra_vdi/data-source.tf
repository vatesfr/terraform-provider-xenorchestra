data "xenorchestra_vdi" "vdi" {
  name_label = "ubuntu-20.04.4-live-server-amd64.iso"
}

resource "xenorchestra_vm" "demo-vm" {
  cdrom = data.xenorchestra_vdi.vdi.id
}
