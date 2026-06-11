# Makefile for Xen Orchestra Terraform Provider v2

.PHONY: import testacc testclient test dist ci docs debug


# Binary name
BINARY_NAME=terraform-provider-xenorchestra

ifdef TF_LOG
    TF_LOG := TF_LOG=$(TF_LOG)
endif

build:
	go build -o $(BINARY_NAME)

debug: 
	go build -o $(BINARY_NAME) -gcflags="all=-N -l"
# Download dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Run acceptance tests (requires TF_ACC=1 and XOA credentials)
testacc:
	TF_ACC=1 go test -v ./...

# Lint the code
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)*

dist:
	./scripts/dist.sh

plan: build
	terraform init
	terraform plan

apply:
	terraform apply

sweep:
	TF_ACC=1 $(TF_LOG) go test $(TEST) -sweep=true -v

# Generate documentation
docs:
	@echo "Generating docs..."
	go generate ./...
	
# Get tools
tools:
	$(GOGET) github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	$(GOGET) golang.org/x/lint/golint
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
	
help:
	@echo "Available targets:"
	@echo "  build          - Build the provider"
	@echo "  docs           - Generate documentation"
	@echo "  deps           - Download dependencies"
	@echo "  test           - Run unit tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  testacc        - Run acceptance tests"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  tools          - Install required tools"
