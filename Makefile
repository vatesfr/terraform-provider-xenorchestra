.PHONY: import testacc testclient test dist

TIMEOUT ?= 40m
GOMAXPROCS ?= 5
RUNNER_TEST ?= TestParallel.*
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

testacc: sweep
	TF_ACC=1 $(TF_LOG) go test $(TEST) -parallel $(GOMAXPROCS) -v -count 1 -timeout $(TIMEOUT)

testparallel:
	go test github.com/ddelnano/terraform-provider-xenorchestra/cmd/testing/parallel -run='TestParallel' -timeout=$(TIMEOUT) -v -count=1

runtestrunner:
	go test github.com/ddelnano/terraform-provider-xenorchestra/cmd/testing/parallel -list='$(RUNNER_TEST)' | go run cmd/testing/main.go
	)
runtestrunnerexample:
	# go test github.com/ddelnano/terraform-provider-xenorchestra/cmd/testing/example -list='.*' | strace -f -e trace=write --signal=none go run cmd/testing/main.go
	go test github.com/ddelnano/terraform-provider-xenorchestra/cmd/testing/example -list='.*' | go run cmd/testing/main.go

runsimple:
	go test github.com/ddelnano/terraform-provider-xenorchestra/cmd/testing/example -list='.*' | go run cmd/testing/simple/main.go
