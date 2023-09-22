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
