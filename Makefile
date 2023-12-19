.PHONY: import testacc testclient test dist ci

TIMEOUT ?= 40m
GOMAXPROCS ?= 5
TF_VERSION ?= v0.14.11
ifdef TEST
    TEST := github.com/ddelnano/terraform-provider-xenorchestra/xoa -run '$(TEST)'
else
    TEST := ./...
endif

ifdef TF_LOG
    TF_LOG := TF_LOG=$(TF_LOG)
endif

build:
	go build -o terraform-provider-xenorchestra

clean:
	rm dist/*

dist:
	./scripts/dist.sh

plan: build
	terraform init
	terraform plan

apply:
	terraform apply

sweep:
	TF_ACC=1 $(TF_LOG) go test $(TEST) -sweep=true -v

test: testclient testacc

testclient:
	cd client; go test $(TEST) -v -count 1

testacc: xoa/testdata/alpine-virt-3.17.0-x86_64.iso
	TF_ACC=1 $(TF_LOG) go test $(TEST) -parallel $(GOMAXPROCS) -v -count 1 -timeout $(TIMEOUT) -sweep=true

# This file was previously stored in the git repo with git lfs. GitHub
# has a very low quota for number of allowed clones and so this needed
# to be removed from the repo. Add a target to enforce that the CI system
# has copied that file into place before the tests run
ci: xoa/testdata/alpine-virt-3.17.0-x86_64.iso
	TF_ACC_TERRAFORM_PATH=/opt/terraform-provider-xenorchestra/bin/$(TF_VERSION) TF_ACC=1 gotestsum --debug --rerun-fails=5 --max-fails=15 --packages=./xoa  -- ./xoa -v -count=1 -timeout=$(TIMEOUT) -sweep=true -parallel=$(GOMAXPROCS)
