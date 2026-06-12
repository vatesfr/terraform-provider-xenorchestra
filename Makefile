# Makefile for Xen Orchestra Terraform Provider v2

.PHONY: fmt lint test testacc build debug clean install generate


# Binary name
BINARY_NAME=terraform-provider-xenorchestra

default: fmt lint install generate

build:
	go build -v -o $(BINARY_NAME)

debug: 
	go build -o $(BINARY_NAME) -gcflags="all=-N -l"

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)*

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

