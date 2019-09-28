.PHONY: import testacc dist

TEST ?= ./...

build:
	GO111MODULE=on go build -o terraform-provider-xenorchestra

clean:
	rm dist/*

dist:
	./scripts/dist.sh

plan: build
	terraform init
	terraform plan

apply:
	terraform apply

testacc:
	TF_ACC=1 go test $(TEST) -v
