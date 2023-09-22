data "xenorchestra_pif" "eth0" {
  device = "eth0"
  vlan   = -1
}

resource "xenorchestra_vm" "demo-vm" {
  // ...
  network {
    network_id = data.xenorchestra_pif.eth0.network
  }
  // ...
}
