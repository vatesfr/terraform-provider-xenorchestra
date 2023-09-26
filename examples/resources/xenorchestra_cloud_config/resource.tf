resource "xenorchestra_cloud_config" "demo" {
  name = "cloud config name"
  template = <<EOF
#cloud-config

runcmd:
 - [ ls, -l, / ]
 - [ sh, -xc, "echo $(date) ': hello world!'" ]
 - [ sh, -c, echo "=========hello world'=========" ]
 - ls -l /root
EOF
}

resource "xenorchestra_vm" "bar" {
  // ...
  cloud_config = xenorchestra_cloud_config.demo.template
  // ...
}
