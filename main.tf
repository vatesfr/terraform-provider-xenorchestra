provider "xenorchestra" {}

resource "xenorchestra_vm" "testing" {
  memoryMax = 1073733632
  CPUs = 1
  cloudConfig = "#cloud-config"
  name_label = ""
  /* template = "" */
}
