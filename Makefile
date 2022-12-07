.PHONY: import testacc testclient test dist

# TODO(ddelnano|): Changes to gotestsum must be upstreamed to handle panic'ed tests.
# Until clone https://github.com/ddelnano/gotestsum and build the binary yourself.
GOTESTSUM_BIN ?= ~/code/gotestsum/gotestsum
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

testacc: sweep
	TF_ACC=1 $(TF_LOG) go test $(TEST) -parallel $(GOMAXPROCS) -v -count 1 -timeout $(TIMEOUT)

ci:
	TF_ACC=1 $(GOTESTSUM_BIN) --rerun-fails=3 --rerun-fails-max-failures=1000 --packages='./...' -- ./... -parallel $(GOMAXPROCS) -v -count=1 -timeout=$(TIMEOUT) -sweep=true
