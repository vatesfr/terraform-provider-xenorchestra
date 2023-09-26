data "xenorchestra_template" "template" {
  name_label = "Ubuntu Bionic Beaver 18.04"
}

data "xenorchestra_sr" "sr" {
  name_label = "Your storage repository label"
}

data "xenorchestra_pif" "eth0" {
  device = "eth0"
  vlan   = -1
}

data "xenorchestra_user" "user" {
  username = "test_user"
}

resource "xenorchestra_resource_set" "rs" {
  name = "new-resource-set"
  subjects = [
    data.xenorchestra_user.user.id,
  ]
  objects = [
    data.xenorchestra_template.template.id,
    data.xenorchestra_sr.sr.id,
    data.xenorchestra_pif.eth0.network,
  ]

  limit {
    type = "cpus"
    quantity = 20
  }

  limit {
    type = "disk"
    quantity = 107374182400
  }

  limit {
    type = "memory"
    quantity = 12884901888
  }
}
