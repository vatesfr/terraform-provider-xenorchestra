data "xenorchestra_host" "host1" {
  name_label = "Your host"
}

# Create a single server network private network
resource "xenorchestra_network" "private_network" {
  name_label = "new network name"
  pool_id = data.xenorchestra_host.host1.pool_id
}

# Create a network with a 22 VLAN tag from the eth0 device
resource "xenorchestra_network" "vlan_network" {
  name_label = "new network name"
  pool_id = data.xenorchestra_host.host1.pool_id
  source_pif_device = "eth0"
  vlan = 22
}
