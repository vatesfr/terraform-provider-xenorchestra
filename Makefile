.PHONY: import testacc

TEST ?= ./...

plan:
	go build -o terraform-provider-xenorchestra
	terraform init
	terraform plan

apply:
	terraform apply

testacc:
	TF_ACC=1 go test $(TEST) -v
