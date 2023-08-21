data "xenorchestra_host" "host1" {
  name_label = "Your host"
}

data "xenorchestra_pif" "pif" {
  device = "eth0"
  vlan = -1
  host_id = data.xenorchestra_host.host1.id
}

resource "xenorchestra_network" "network" {
  name_label = "new network name"
  pool_id = data.xenorchestra_host.host1.pool_id
  pif_id = data.xenorchestra_pif.pif.id
  vlan = 22
}
