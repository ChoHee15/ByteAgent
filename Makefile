.PHONY: test-unit test test-integration test-all

TEST_PKGS := $(shell go list ./... | grep -v '/playground$$')

test-unit:
	@echo "Running default mock/local tests..."
	go test $(TEST_PKGS) -v -race -cover

test: test-unit

test-integration:
	@echo "Running real API smoke tests..."
	go test $(TEST_PKGS) -tags=integration -run Integration -v -count=1

test-all: test test-integration
