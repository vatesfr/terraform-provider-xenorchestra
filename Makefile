.PHONY: import testacc testclient test dist

TIMEOUT ?= 40m
GOMAXPROCS ?= 5
ifdef TEST
    TEST := ./... -run '$(TEST)'
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

testacc:
	TF_ACC=1 $(TF_LOG) go test $(TEST) -parallel $(GOMAXPROCS) -v -count 1 -timeout $(TIMEOUT) -sweep=true
ci:
	TF_ACC=1 gotestsum --debug --rerun-fails=5 --max-fails=15 --packages=./xoa  -- ./xoa -v -count=1 -timeout=$(TIMEOUT) -sweep=true -parallel=$(GOMAXPROCS)
