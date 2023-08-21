data "xenorchestra_network" "net" {
  name_label = "Pool-wide network associated with eth0"
}

resource "xenorchestra_vm" "demo-vm" {
  // ...
  network {
    network_id = data.xenorchestra_network.net.id
  }
}

