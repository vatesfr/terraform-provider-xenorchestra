# Terraform Provider for Xen Orchestra

## Status

This is an experimental terraform provider for [Xenorchestra](https://github.com/vatesfr/xen-orchestra).

## Example Use

```hcl
# You must set the following environment variables: XOA_HOST, XOA_USER and XOA_PASSWORD
provider "xenorchestra" {}

resource "xenorchestra_cloud_config" "bar" {
    name = "name"
    template = <<EOF
#cloud-init

# Add cloud init configuration
EOF
}

data "xenorchestra_template" "template" {
    name_label = "Ubuntu Bionic Beaver 18.04"
}

data "xenorchestra_pif" "pif" {
    device = "eth1"
}

resource "xenorchestra_vm" "bar" {
    memory_max = 1073733632
    cpus  = 1
    cloud_config = "${xenorchestra_cloud_config.bar.template}"
    name_label = "Name"
    name_description = "description"
    template = "${data.xenorchestra_template.template.id}"
    network {
	network_id = "${data.xenorchestra_pif.pif.network}"
    }

    disk {
      sr_id = "7f469400-4a2b-5624-cf62-61e522e50ea1"
      name_label = "Ubuntu Bionic Beaver 18.04_imavo"
      size = 32212254720 
    }
}
```
