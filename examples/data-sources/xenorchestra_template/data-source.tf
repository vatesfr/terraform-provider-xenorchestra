data "xenorchestra_template" "template" {
  name_label = "Ubuntu Bionic Beaver 18.04"
}

resource "xenorchestra_vm" "demo-vm" {
  // ...
  template = data.xenorchestra_template.template.id
  // ...
}
