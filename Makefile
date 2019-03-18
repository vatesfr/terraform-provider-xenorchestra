.PHONY: import testacc

TEST ?= ./...

import:
	go build -o terraform-provider-xenorchestra
	terraform init
	terraform import xenorchestra_vm.testing 77c6637c-fa3d-0a46-717e-296208c40169
plan:
	go build -o terraform-provider-xenorchestra
	terraform init
	terraform plan

apply:
	terraform apply

testacc:
	TF_ACC=1 go test $(TEST) -v
