/*
# cloud_config.tftpl file used by the cloudinit templating.

#cloud-config
hostname: ${hostname}
fqdn: ${hostname}.${domain}
package_upgrade: true
*/

# Content of the terraform files
data "xenorchestra_pool" "pool" {
  name_label = "pool name"
}

data "xenorchestra_template" "template" {
  name_label = "Puppet agent - Bionic 18.04 - 1"
}

data "xenorchestra_network" "net" {
  name_label = "Pool-wide network associated with eth0"
}

resource "xenorchestra_cloud_config" "bar" {
  name = "cloud config name"
  # Template the cloudinit if needed
  template = templatefile("cloud_config.tftpl", {
    hostname = "your-hostname"
    domain   = "your.domain.com"
  })
}

resource "xenorchestra_vm" "bar" {
  memory_max       = 1073733632
  cpus             = 1
  cloud_config     = xenorchestra_cloud_config.bar.template
  name_label       = "Name"
  name_description = "description"
  template         = data.xenorchestra_template.template.id

  # Prefer to run the VM on the primary pool instance
  affinity_host = data.xenorchestra_pool.pool.master
  network {
    network_id = data.xenorchestra_network.net.id
  }

  disk {
    sr_id      = "7f469400-4a2b-5624-cf62-61e522e50ea1"
    name_label = "Ubuntu Bionic Beaver 18.04_imavo"
    size       = 32212254720
  }

  tags = [
    "Ubuntu",
    "Bionic",
  ]

  // Override the default create timeout from 5 mins to 20.
  timeouts {
    create = "20m"
  }

  // Note: Xen Orchestra populates values within Xenstore and will need ignored via
  // lifecycle ignore_changes or modeled in your terraform code
  xenstore = {
    key1 = "val1"
    key2 = "val2"
  }
}

# vm resource that waits until its first network interface
# is assigned an IP via DHCP
resource "xenorchestra_vm" "vm" {
  # Specify VM with two network interfaces
  network {
    network_id       = "7ed8998b-405c-40b5-b164-f9058efcf6b4"
    expected_ip_cidr = "10.0.0.0/16"
  }
  network {
    network_id = "b4cf8532-ae43-493b-9fc6-a6b456d16876"
  }
  # required arguments for xenorchestra_vm
  cpus = 2
  disk {
    sr_id      = "7f469400-4a2b-5624-cf62-61e522e50ea1"
    name_label = "Ubuntu Bionic Beaver 18.04_imavo"
    size       = 32212254720
  }
  memory_max = 1073733632
  name_label = "Name"
  template   = data.xenorchestra_template.template.id
}

output "first-network-interface-ips" {
  value = xenorchestra_vm.vm.network[0].ipv4_addresses
}
