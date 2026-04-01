.PHONY: test-unit test test-integration test-all build package clean

TEST_PKGS := $(shell go list ./... | grep -v '/playground$$')
BINARY_NAME ?= code-agent
DIST_DIR ?= dist
GOOS ?= linux
GOARCH ?= amd64
VERSION ?= dev
COMMIT ?= unknown

test-unit:
	@echo "Running default mock/local tests..."
	go test $(TEST_PKGS) -v -race -cover

test: test-unit

test-integration:
	@echo "Running real API smoke tests..."
	go test $(TEST_PKGS) -tags=integration -run Integration -v -count=1

test-all: test test-integration

build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./cmd/code-agent

package: build
	@echo "Packaging $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	tar -C $(DIST_DIR) -czf $(DIST_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH).tar.gz $(BINARY_NAME)-$(GOOS)-$(GOARCH)

clean:
	rm -rf $(DIST_DIR)
