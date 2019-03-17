provider "xenorchestra" {}

resource "xenorchestra_vm" "testing" {
  memory_max = 1073733632
  cpus = 1
  cloud_config = ""
  name_label = "Hello from terraform!"
  name_description = "Testingsdfsdf"
  template = ""
}

resource "xenorchestra_cloud_config" "test_config" {
    name = "testing"
    template = <<EOF
#cloud-init

packages:
- build-essentials
EOF
}

/*
vm.create name_label="vm name" bootAfterCreate=true cloudConfig='a29bf6bc-8e15-4202-9fe5-c86884b6e08e' template='2dd0373e-0ed5-7413-a57f-1958d03b698c'
*/
