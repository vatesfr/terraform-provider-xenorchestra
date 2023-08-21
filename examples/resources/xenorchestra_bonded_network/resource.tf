data "xenorchestra_host" "host1" {
  name_label = "Your host"
}

data "xenorchestra_pif" "eth1" {
  device = "eth1"
  vlan = -1
  host_id = data.xenorchestra_host.host1.id
}

data "xenorchestra_pif" "eth2" {
  device = "eth2"
  vlan = -1
  host_id = data.xenorchestra_host.host1.id
}

resource "xenorchestra_network" "network" {
  name_label = "new network name"
  bond_mode = "active-backup"
  pool_id = data.xenorchestra_host.host1.pool_id
  pif_ids = [
    data.xenorchestra_pif.eth1.id,
    data.xenorchestra_pif.eth2.id,
  ]
}
