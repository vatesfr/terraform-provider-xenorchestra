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

# Create a bonded network from normal PIFs
resource "xenorchestra_bonded_network" "network" {
  name_label = "new network name"
  bond_mode = "active-backup"
  pool_id = data.xenorchestra_host.host1.pool_id
  pif_ids = [
    data.xenorchestra_pif.eth1.id,
    data.xenorchestra_pif.eth2.id,
  ]
}

# Create a bonded network from PIFs on VLANs
data "xenorchestra_pif" "eth1_vlan" {
  device = "eth1"
  vlan = 15
  host_id = data.xenorchestra_host.host1.id
}

data "xenorchestra_pif" "eth2_vlan" {
  device = "eth2"
  vlan = 15
  host_id = data.xenorchestra_host.host1.id
}

# Create a bonded network from normal PIFs
resource "xenorchestra_bonded_network" "network_vlan" {
  name_label = "new network name"
  bond_mode = "active-backup"
  pool_id = data.xenorchestra_host.host1.pool_id
  pif_ids = [
    data.xenorchestra_pif.eth1_vlan.id,
    data.xenorchestra_pif.eth2_vlan.id,
  ]
}
