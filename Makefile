# Makefile for Xen Orchestra Terraform Provider v2

.PHONY: all build test generate clean tools

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary name
BINARY_NAME=terraform-provider-xenorchestra

# Build the provider
build:
	$(GOBUILD) -o $(BINARY_NAME) .

# Build for Linux AMD64 (common for testing)
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_linux_amd64 .

# Build for macOS ARM64
build-darwin:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINARY_NAME)_darwin_arm64 .

# Build for Windows AMD64
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_windows_amd64.exe .

# Build all platforms
build-all: build-linux build-darwin build-windows

# Generate documentation
generate:
	$(GOBUILD) -o tfplugindocs-generated ./tools/tfplugindocs
	./tfplugindocs-generated generate --provider-name xenorchestra

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -cover ./...

# Run acceptance tests (requires TF_ACC=1 and XOA credentials)
testacc:
	TF_ACC=1 $(GOTEST) -v ./...

# Lint the code
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)*

# Get tools
tools:
	$(GOGET) github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	$(GOGET) golang.org/x/lint/golint
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint

# All targets
all: deps build

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the provider"
	@echo "  build-linux    - Build for Linux AMD64"
	@echo "  build-darwin   - Build for macOS ARM64"
	@echo "  build-windows  - Build for Windows AMD64"
	@echo "  build-all      - Build for all platforms"
	@echo "  generate       - Generate documentation"
	@echo "  deps           - Download dependencies"
	@echo "  test           - Run unit tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  testacc        - Run acceptance tests"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  tools          - Install required tools"
