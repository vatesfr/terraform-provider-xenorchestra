data "xenorchestra_host" "host1" {
  name_label = "Your host"
}
resource "xenorchestra_vm" "node" {
    //...
    affinity_host = data.xenorchestra_host.host1.id
    //...
}
