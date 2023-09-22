# cloud_config.tftpl file used by the cloudinit templating.

#cloud-config
hostname: ${hostname}
fqdn: ${hostname}.${domain}
package_upgrade: true
```

```hcl
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
    domain = "your.domain.com"
  })
}

resource "xenorchestra_vm" "bar" {
    memory_max = 1073733632
    cpus  = 1
    cloud_config = xenorchestra_cloud_config.bar.template
    name_label = "Name"
    name_description = "description"
    template = data.xenorchestra_template.template.id

    # Prefer to run the VM on the primary pool instance
    affinity_host = data.xenorchestra_pool.pool.master
    network {
      network_id = data.xenorchestra_network.net.id
    }

    disk {
      sr_id = "7f469400-4a2b-5624-cf62-61e522e50ea1"
      name_label = "Ubuntu Bionic Beaver 18.04_imavo"
      size = 32212254720
    }

    tags = [
      "Ubuntu",
      "Bionic",
    ]

    // Override the default create timeout from 5 mins to 20.
    timeouts {
      create = "20m"
    }
}
